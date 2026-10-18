package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Conf RabbitMQ 配置
type Conf struct {
	Url             string // amqp://guest:guest@127.0.0.1:5672/
	ChatExchange    string // 上行消息交换机（direct）
	PushExchange    string // 下行投递交换机（direct）
	ChatQueue       string // 持久化消费者的共享队列
	PushQueuePrefix string // 每个网关节点专属队列前缀
	CanalQueue      string `json:",optional"` // Canal(binlog) 变更事件队列；由 Canal 中间件声明并投递，本服务只消费。仅 rpc 需要，网关可不配置
}

const (
	ChatRoutingKey = "chat" // 上行统一路由键
	// NodeRKPrefix 下行节点级路由键前缀：node.<nodeID>。
	// 网关按「用户->节点」路由表把消息投到目标节点，由网关拆信封后投递给具体连接。
	NodeRKPrefix = "node."
)

// Client 封装 RabbitMQ 连接、发布与订阅
type Client struct {
	conf   Conf
	nodeID string

	conn  *amqp.Connection
	pub   *amqp.Channel
	pushQ string
}

// NewClient 建立连接、声明两个交换机，并声明本节点专属的推送队列
func NewClient(conf Conf, nodeID string) (*Client, error) {
	conn, err := amqp.Dial(conf.Url)
	if err != nil {
		return nil, err
	}
	pub, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	if err = pub.ExchangeDeclare(conf.ChatExchange, "direct", true, false, false, false, nil); err != nil {
		return nil, err
	}
	if err = pub.ExchangeDeclare(conf.PushExchange, "direct", true, false, false, false, nil); err != nil {
		return nil, err
	}

	c := &Client{
		conf:   conf,
		nodeID: nodeID,
		conn:   conn,
		pub:    pub,
	}
	// 本节点专属推送队列：非持久、自动删除、独占
	// 注：RabbitMQ 4.x 已默认禁止“非持久且非独占”的队列，故必须 exclusive=true
	// 只有网关（api）需要接收下行推送；业务服务（rpc）传空 nodeID，不声明该队列
	if nodeID != "" {
		c.pushQ = fmt.Sprintf("%s.%s", conf.PushQueuePrefix, nodeID)
		if _, err = pub.QueueDeclare(c.pushQ, false, true, true, false, nil); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// PublishChat 上行消息：客户端 -> 聊天持久化消费者
func (c *Client) PublishChat(body []byte) error {
	return c.pub.PublishWithContext(context.Background(), c.conf.ChatExchange, ChatRoutingKey, false, false,
		amqp.Publishing{ContentType: "application/json", Body: body})
}

// PushEnvelope 下行投递信封：To 为目标用户 uuid，Body 为原样写给浏览器的报文。
// 路由键改为节点级后不再携带用户信息，故需信封指明具体收件人。
type PushEnvelope struct {
	To   string          `json:"to"`
	Body json.RawMessage `json:"body"`
}

// BindNode 启动时把本节点队列绑定到节点级路由键 node.<nodeID>（只绑一次）。
func (c *Client) BindNode(nodeID string) error {
	if c.pushQ == "" || nodeID == "" {
		return nil
	}
	return c.pub.QueueBind(c.pushQ, NodeRKPrefix+nodeID, c.conf.PushExchange, false, nil)
}

// PublishPushTo 把信封投递到指定网关节点（路由键 node.<nodeID>）
func (c *Client) PublishPushTo(nodeID string, body []byte) error {
	if nodeID == "" {
		return nil
	}
	return c.pub.PublishWithContext(context.Background(), c.conf.PushExchange, NodeRKPrefix+nodeID, false, false,
		amqp.Publishing{ContentType: "application/json", Body: body})
}

// ConsumePush 消费本节点推送队列（跨节点投递的落地点）
func (c *Client) ConsumePush(handler func(routingKey string, body []byte)) error {
	if c.pushQ == "" {
		return nil
	}
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}
	msgs, err := ch.Consume(c.pushQ, "", true, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		for m := range msgs {
			handler(m.RoutingKey, m.Body)
		}
	}()
	return nil
}

// ConsumeChat 消费上行消息（共享队列，人工 ack，保证不丢）
func (c *Client) ConsumeChat(handler func(body []byte) error) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}
	if _, err = ch.QueueDeclare(c.conf.ChatQueue, true, false, false, false, nil); err != nil {
		return err
	}
	// 关键：把上行队列绑定到 chat 交换机，否则 PublishChat 发出的消息没有队列接收，会被直接丢弃
	if err = ch.QueueBind(c.conf.ChatQueue, ChatRoutingKey, c.conf.ChatExchange, false, nil); err != nil {
		return err
	}
	if err = ch.Qos(64, 0, false); err != nil {
		return err
	}
	msgs, err := ch.Consume(c.conf.ChatQueue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		for m := range msgs {
			if err := handler(m.Body); err != nil {
				_ = m.Nack(false, true) // 处理失败，重新入队
				continue
			}
			_ = m.Ack(false)
		}
	}()
	return nil
}

// ConsumeCanal 消费 Canal(binlog) 变更事件队列。
// 队列由 Canal 中间件侧声明与投递，这里只消费、不声明，避免与其队列参数冲突；
// 启动时若队列尚未就绪会自动重试（Canal 容器通常晚于/先于 RPC 就绪均可能发生）。
func (c *Client) ConsumeCanal(queue string, handler func(body []byte) error) {
	if queue == "" {
		return
	}
	go func() {
		for {
			ch, err := c.conn.Channel()
			if err != nil {
				logx.Errorf("Canal 消费者打开信道失败，稍后重试：%v", err)
				time.Sleep(3 * time.Second)
				continue
			}
			if err = ch.Qos(64, 0, false); err != nil {
				_ = ch.Close()
				logx.Errorf("Canal 消费者设置 Qos 失败：%v", err)
				time.Sleep(3 * time.Second)
				continue
			}
			msgs, err := ch.Consume(queue, "", false, false, false, false, nil)
			if err != nil {
				// 多为队列尚未创建，等待后重试
				_ = ch.Close()
				logx.Errorf("Canal 队列 %s 暂不可消费，稍后重试：%v", queue, err)
				time.Sleep(3 * time.Second)
				continue
			}
			for m := range msgs {
				if err := handler(m.Body); err != nil {
					// 处理失败不重新入队，避免毒消息死循环（缓存本身有 TTL 兜底）
					logx.Errorf("处理 Canal 事件失败：%v", err)
					_ = m.Nack(false, false)
					continue
				}
				_ = m.Ack(false)
			}
			// 信道被服务端关闭后重建
			_ = ch.Close()
			time.Sleep(time.Second)
		}
	}()
}

func (c *Client) Close() {
	if c.pub != nil {
		_ = c.pub.Close()
	}
	if c.conn != nil {
		_ = c.conn.Close()
	}
}
