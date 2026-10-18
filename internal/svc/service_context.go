package svc

import (
	"fmt"
	"os"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"
	"github.com/zeromicro/go-zero/zrpc"

	"echochat/internal/config"
	"echochat/internal/ws"
	"echochat/pkg/es"
	"echochat/pkg/mq"
	"echochat/pkg/redisx"
	"echochat/pkg/registry"
	"echochat/rpc/core_client"
)

type ServiceContext struct {
	Config   config.Config
	Core     core_client.Core
	Mq       *mq.Client
	Hub      *ws.Hub
	Es       *es.Client
	NodeID   string
	Registry *registry.Registry
}

// nodeID 生成稳定的网关节点标识：env NODE_ID 优先，其次配置项，最后回退 hostname-port。
// 这里用 port 而不是 pid——重启后队列名 push.gw.<id> 与路由表 value 都能保持稳定。
func nodeID(c config.Config) string {
	if v := os.Getenv("NODE_ID"); v != "" {
		return v
	}
	if c.NodeID != "" {
		return c.NodeID
	}
	host, _ := os.Hostname()
	return fmt.Sprintf("%s-%d", host, c.Port)
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 网关用 Redis 校验登录令牌（RBAC 鉴权）与维护「用户->节点」路由表
	redisx.Init(redisx.Conf{Host: c.RedisConf.Host, Port: c.RedisConf.Port, Password: c.RedisConf.Password, Db: c.RedisConf.Db})
	cli := zrpc.MustNewClient(c.CoreRpc)

	id := nodeID(c)
	m, err := mq.NewClient(c.RabbitMQ, id)
	if err != nil {
		panic(err)
	}
	// 把本节点推送队列绑定到节点级路由键 node.<nodeID>，接收路由到本节点的下行消息
	if err = m.BindNode(id); err != nil {
		panic(err)
	}

	// etcd 节点注册：让集群知道有哪些网关存活，并在节点消失时触发路由清理（未配置则跳过）
	reg, rerr := registry.New(c.Registry)
	if rerr != nil {
		logx.Errorf("etcd 注册客户端初始化失败：%v", rerr)
	}
	if reg != nil {
		if err = reg.Register(id, fmt.Sprintf("%s:%d", c.Host, c.Port)); err != nil {
			logx.Errorf("网关节点注册到 etcd 失败：%v", err)
		}
		proc.AddShutdownListener(func() { reg.Unregister() })
	}

	return &ServiceContext{
		Config:   c,
		Core:     core_client.NewCore(cli),
		Mq:       m,
		Hub:      ws.NewHub(m, c.WsAllowedOrigins, id),
		Es:       es.New(c.Elastic),
		NodeID:   id,
		Registry: reg,
	}
}
