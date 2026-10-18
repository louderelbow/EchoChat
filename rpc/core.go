package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"echochat/rpc/core"
	"echochat/rpc/internal/canal"
	"echochat/rpc/internal/chat"
	"echochat/rpc/internal/config"
	"echochat/rpc/internal/server"
	"echochat/rpc/internal/svc"

	"echochat/pkg/registry"
	"echochat/pkg/route"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/core.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 敏感凭据允许用环境变量覆盖；未设置时沿用配置文件值，保证本地开发可直接运行
	if v := os.Getenv("ECHO_MYSQL_PASSWORD"); v != "" {
		c.Mysql.Password = v
	}
	if v := os.Getenv("ECHO_RABBITMQ_URL"); v != "" {
		c.RabbitMQ.Url = v
	}

	ctx := svc.NewServiceContext(c)

	// 初始化 ES 索引（未配置 ES 时为空操作）
	if err := ctx.Es.EnsureIndex(context.Background()); err != nil {
		fmt.Printf("ES 索引初始化失败（聊天记录检索可能不可用）：%v\n", err)
	}

	// 启动上行消息消费者（RabbitMQ）：持久化 + 跨节点投递
	if err := chat.Start(ctx); err != nil {
		panic(err)
	}

	// 启动 Canal(binlog) 缓存失效消费者：消费 Canal 投递到 RabbitMQ 的变更事件，
	// 精准删除对应 Redis 缓存（替代原先业务代码中的手动失效）
	canal.Start(ctx)

	// 监听 etcd 网关节点上下线：节点消失（崩溃/退出）时立即清理其残留路由，
	// 把故障恢复从「等路由 TTL 过期」降到「秒级」（未配置 etcd 时为空操作）
	reg, rerr := registry.New(c.Registry)
	if rerr != nil {
		fmt.Printf("etcd 注册客户端初始化失败（网关节点清理可能不可用）：%v\n", rerr)
	}
	if reg != nil {
		reg.WatchNodes(func(added, removed []string) {
			if len(added) > 0 {
				fmt.Printf("网关节点上线：%v\n", added)
			}
			for _, node := range removed {
				n, derr := route.DeleteByNode(node)
				if derr != nil {
					fmt.Printf("清理网关节点 %s 的残留路由失败：%v\n", node, derr)
					continue
				}
				fmt.Printf("网关节点 %s 下线，已清理 %d 条残留路由\n", node, n)
			}
		})
	}

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		core.RegisterCoreServer(grpcServer, server.NewCoreServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
