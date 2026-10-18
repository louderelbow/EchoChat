package ws

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/metric"

	"echochat/pkg/constants"
	"echochat/pkg/mq"
	"echochat/pkg/route"
)

// 允许连接 WebSocket 的 Origin 主机名白名单；为空时仅允许本机来源（localhost/127.0.0.1）。
// 由 NewHub 注入，生产环境通过配置 WsAllowedOrigins 填写实际域名。
var allowedOriginHosts []string

// 网关连接层指标（注册进 go-zero 内置 Prometheus registry；只加业务序列，刮取/看板交给观测栈）
var (
	connsGauge = metric.NewGaugeVec(&metric.GaugeVecOpts{
		Namespace: "im", Subsystem: "gateway", Name: "connections",
		Help: "本网关节点当前 WebSocket 连接数", Labels: []string{"node"},
	})
	pushTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: "im", Subsystem: "gateway", Name: "push_total",
		Help: "下行推送成功写入本地连接的次数",
	})
	pushDroppedTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: "im", Subsystem: "gateway", Name: "push_dropped_total",
		Help: "下行推送被丢弃的次数（目标不在本节点或发送缓冲已满）",
	})
)

// originHost 解析 Origin 头里的主机名（忽略端口）。
func originHost(origin string) string {
	if origin == "" {
		return ""
	}
	u, err := url.Parse(origin)
	if err != nil {
		return ""
	}
	return u.Hostname()
}

// checkOrigin 校验 WebSocket 握手的 Origin：无 Origin（非浏览器客户端）放行，
// 有 Origin 则要求主机名在白名单内，防止任意站点跨站连接。
func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	host := originHost(origin)
	allowed := allowedOriginHosts
	if len(allowed) == 0 {
		allowed = []string{"localhost", "127.0.0.1"}
	}
	for _, a := range allowed {
		if strings.EqualFold(a, host) {
			return true
		}
	}
	return false
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
	CheckOrigin:     checkOrigin,
}

type client struct {
	conn *websocket.Conn
	uuid string
	send chan []byte
	done chan struct{}
	once sync.Once
}

// Hub 管理本网关节点上的在线连接（进程内），并借助 RabbitMQ + 路由表完成跨节点投递。
// 同一用户可以同时存在多条连接（多个标签页/客户端），互不踢线、都能收到下行消息。
type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*client]struct{}
	mq      *mq.Client
	nodeID  string
}

// NewHub 创建网关连接管理器，并启动路由表续期循环
func NewHub(m *mq.Client, allowedOrigins []string, nodeID string) *Hub {
	allowedOriginHosts = allowedOrigins
	h := &Hub{
		clients: make(map[string]map[*client]struct{}),
		mq:      m,
		nodeID:  nodeID,
	}
	go h.renewLoop()
	return h
}

// Serve 处理 /wss?client_id=<uuid> 的升级、上线登记与心跳设置
func (h *Hub) Serve(w http.ResponseWriter, r *http.Request, clientID string) error {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	c := &client{
		conn: conn,
		uuid: clientID,
		send: make(chan []byte, constants.CHANNEL_SIZE),
		done: make(chan struct{}),
	}
	// 心跳：设定读取截止时间，收到 Pong 就刷新；浏览器会自动回协议级 Pong，无需前端改动
	_ = conn.SetReadDeadline(time.Now().Add(constants.WS_PONG_WAIT))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(constants.WS_PONG_WAIT))
	})

	// 同一用户允许多条连接并存（多标签页/多客户端），新连接不再踢掉旧连接，
	// 避免两个客户端互相踢线形成 3 秒一轮的重连风暴、进而丢消息。
	h.mu.Lock()
	if h.clients[clientID] == nil {
		h.clients[clientID] = make(map[*client]struct{})
	}
	h.clients[clientID][c] = struct{}{}
	count := h.connCountLocked()
	h.mu.Unlock()
	connsGauge.Set(float64(count), h.nodeID)

	// 登记「用户->节点」路由表（跨节点 last-writer-wins）：下行据此找到本节点
	if err := route.Set(clientID, h.nodeID, route.RouteTTL); err != nil {
		logx.Errorf("写入用户路由失败（Redis 异常，跨节点投递可能受影响）: %v", err)
	}

	_ = conn.WriteMessage(websocket.TextMessage, []byte("欢迎来到 EchoChat 聊天服务器"))
	go h.readLoop(c)
	go h.writeLoop(c)
	return nil
}

// Remove 主动断开某用户在本节点的全部连接（如退出登录）。跨节点场景由路由表 + 节点侧各自处理。
func (h *Hub) Remove(uuid string) {
	h.mu.RLock()
	set := h.clients[uuid]
	conns := make([]*client, 0, len(set))
	for c := range set {
		conns = append(conns, c)
	}
	h.mu.RUnlock()
	for _, c := range conns {
		h.remove(c)
	}
}

// connCountLocked 统计当前连接总数（调用方需已持有 h.mu）
func (h *Hub) connCountLocked() int {
	n := 0
	for _, set := range h.clients {
		n += len(set)
	}
	return n
}

// remove 下线单条连接：从该用户的连接集合里移除自己。
// 只有当该用户最后一条连接也断开时，才清理「用户->节点」路由表项，
// 避免多客户端场景下某个客户端断开就把仍在线用户的路径误删。
func (h *Hub) remove(c *client) {
	h.mu.Lock()
	removed := false
	last := false
	if set, ok := h.clients[c.uuid]; ok {
		if _, exists := set[c]; exists {
			delete(set, c)
			removed = true
		}
		if len(set) == 0 {
			delete(h.clients, c.uuid)
			last = true
		}
	}
	count := h.connCountLocked()
	h.mu.Unlock()

	connsGauge.Set(float64(count), h.nodeID)

	c.close()
	if removed && last {
		// CAS 删：仅当路由仍指向本节点时才删，避免误删已被其他节点抢占的路由
		if _, err := route.CompareAndDel(c.uuid, h.nodeID); err != nil {
			logx.Errorf("清理用户路由失败: %v", err)
		}
	}
}

// Push 把消息推给本节点上该用户的全部在线连接（多标签页/多客户端都能收到）
func (h *Hub) Push(uuid string, msg []byte) {
	h.mu.RLock()
	set := h.clients[uuid]
	conns := make([]*client, 0, len(set))
	for c := range set {
		conns = append(conns, c)
	}
	h.mu.RUnlock()
	if len(conns) == 0 {
		// 路由表可能已过期（用户已迁到别的节点或已下线）
		pushDroppedTotal.Inc()
		return
	}
	for _, c := range conns {
		select {
		case c.send <- msg:
			pushTotal.Inc()
		case <-c.done:
			pushDroppedTotal.Inc()
		default:
			// 发送缓冲已满：丢弃该条，避免阻塞整个投递链路
			pushDroppedTotal.Inc()
		}
	}
}

// OnPush 消费本节点队列的入口：路由键形如 node.<nodeID>，载荷为 PushEnvelope。
// 拆出信封里的 To 与 Body 后，只把 Body 原样写给浏览器（前端收到的内容与改造前一致）。
func (h *Hub) OnPush(routingKey string, raw []byte) {
	if routingKey != mq.NodeRKPrefix+h.nodeID {
		return
	}
	var env mq.PushEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		logx.Errorf("解析下行信封失败: %v", err)
		return
	}
	h.Push(env.To, env.Body)
}

// renewLoop 周期性为本地连接续期路由表（TTL 即「在线」信号）
func (h *Hub) renewLoop() {
	ticker := time.NewTicker(constants.ROUTE_RENEW_INTERVAL)
	defer ticker.Stop()
	for range ticker.C {
		h.mu.RLock()
		uuids := make([]string, 0, len(h.clients))
		for u := range h.clients {
			uuids = append(uuids, u)
		}
		h.mu.RUnlock()
		if len(uuids) == 0 {
			continue
		}
		if err := route.Renew(uuids, h.nodeID, route.RouteTTL); err != nil {
			logx.Errorf("续期用户路由失败: %v", err)
		}
	}
}

func (h *Hub) readLoop(c *client) {
	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			// 读超时（心跳失联）或对端关闭，统一清理
			h.remove(c)
			return
		}
		// 上行消息统一投到 im.chat，由核心服务持久化后再路由下发
		if err := h.mq.PublishChat(data); err != nil {
			logx.Errorf("上行消息投递失败（MQ 可能已断开，需要重启服务）: %v", err)
		}
	}
}

func (h *Hub) writeLoop(c *client) {
	ticker := time.NewTicker(constants.WS_PING_PERIOD)
	defer ticker.Stop()
	for {
		select {
		case msg := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				h.remove(c)
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				h.remove(c)
				return
			}
		case <-c.done:
			return
		}
	}
}

func (c *client) close() {
	c.once.Do(func() { close(c.done) })
	_ = c.conn.Close()
}
