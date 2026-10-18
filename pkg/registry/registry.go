package registry

import (
	"context"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Conf etcd 服务注册配置；EtcdEndpoints 为空表示不启用（本地无 etcd 时网关仍可单机运行）
type Conf struct {
	EtcdEndpoints []string // 例如 ["127.0.0.1:2379"]
	Prefix        string   // 注册根路径，默认 im
}

// leaseTTL 租约时长（秒）：网关崩溃后 etcd 最多在此时间内回收该节点 key
const leaseTTL = 15

// Registry 基于 etcd 租约的网关节点注册表。
// 目的：① 让集群知道「有哪些网关节点存活」；② 节点消失时通过 Watch 触发路由快速清理。
type Registry struct {
	cli    *clientv3.Client
	prefix string
	// nodeID/leaseID 记录自身注册，便于优雅退出
	nodeID  string
	leaseID clientv3.LeaseID
	cancel  context.CancelFunc
}

// New 创建注册客户端；未配置 endpoints 时返回 nil（调用方无需判空，方法均已做 nil 保护）
func New(conf Conf) (*Registry, error) {
	if len(conf.EtcdEndpoints) == 0 {
		return nil, nil
	}
	prefix := conf.Prefix
	if prefix == "" {
		prefix = "im"
	}
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   conf.EtcdEndpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return &Registry{cli: cli, prefix: prefix}, nil
}

func (r *Registry) gatewayKey(nodeID string) string {
	return r.prefix + "/gateways/" + nodeID
}

func (r *Registry) gatewayPrefix() string {
	return r.prefix + "/gateways/"
}

// Register 以「租约 + 自动续约」注册本网关节点，addr 为对外可达地址。
// 进程崩溃后租约过期，key 自动消失，Watch 端即可感知。
func (r *Registry) Register(nodeID, addr string) error {
	if r == nil {
		return nil
	}
	ctx := context.Background()
	lease, err := r.cli.Grant(ctx, leaseTTL)
	if err != nil {
		return err
	}
	if _, err = r.cli.Put(ctx, r.gatewayKey(nodeID), addr, clientv3.WithLease(lease.ID)); err != nil {
		return err
	}
	kaCtx, cancel := context.WithCancel(context.Background())
	ch, err := r.cli.KeepAlive(kaCtx, lease.ID)
	if err != nil {
		cancel()
		return err
	}
	// 消费续约响应，保持租约存活
	go func() {
		for range ch {
		}
	}()
	r.nodeID = nodeID
	r.leaseID = lease.ID
	r.cancel = cancel
	logx.Infof("网关节点已注册到 etcd：%s -> %s", nodeID, addr)
	return nil
}

// Unregister 优雅退出：撤销租约（key 立即消失），关闭连接
func (r *Registry) Unregister() {
	if r == nil {
		return
	}
	if r.cancel != nil {
		r.cancel()
	}
	ctx := context.Background()
	if r.leaseID != 0 {
		_, _ = r.cli.Revoke(ctx, r.leaseID)
	}
	_ = r.cli.Close()
}

// WatchNodes 监听网关目录变化：节点上线回调 added，节点消失（含租约过期/优雅退出）回调 removed。
// rpc 侧用 removed 触发路由表清理。
func (r *Registry) WatchNodes(onChange func(added, removed []string)) {
	if r == nil {
		return
	}
	go func() {
		watchCh := r.cli.Watch(context.Background(), r.gatewayPrefix(), clientv3.WithPrefix())
		for resp := range watchCh {
			var added, removed []string
			for _, ev := range resp.Events {
				id := strings.TrimPrefix(string(ev.Kv.Key), r.gatewayPrefix())
				if id == "" {
					continue
				}
				switch ev.Type {
				case clientv3.EventTypePut:
					added = append(added, id)
				case clientv3.EventTypeDelete:
					removed = append(removed, id)
				}
			}
			if len(added) > 0 || len(removed) > 0 {
				onChange(added, removed)
			}
		}
	}()
}
