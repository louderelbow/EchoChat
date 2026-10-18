package config

import (
	"github.com/zeromicro/go-zero/zrpc"
	"echochat/pkg/db"
	"echochat/pkg/es"
	"echochat/pkg/mq"
	"echochat/pkg/redisx"
	"echochat/pkg/registry"
)

type StaticConf struct {
	AvatarPath string
	FilePath   string
}

type Config struct {
	zrpc.RpcServerConf
	Mysql     db.MysqlConf
	RedisConf redisx.Conf
	RabbitMQ  mq.Conf
	Static    StaticConf
	// Elastic 聊天记录全文检索（可选，未配置则不启用）
	Elastic es.Conf
	// Registry etcd 网关注册配置：rpc 侧监听节点上下线并做路由表清理（可选）
	Registry registry.Conf `json:",optional"`
}
