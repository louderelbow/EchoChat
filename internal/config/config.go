package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"echochat/pkg/es"
	"echochat/pkg/mq"
	"echochat/pkg/redisx"
	"echochat/pkg/registry"
)

type StaticConf struct {
	AvatarPath string
	FilePath   string
}

// OssConf 阿里云 OSS 存储配置；AK/SK 由环境变量 ${ACID}/${ACSC} 注入。
type OssConf struct {
	Endpoint         string
	BucketName       string
	AccessKeyID      string
	AccessKeySecret  string
	AutoCreateBucket bool
}

type Config struct {
	rest.RestConf
	RabbitMQ mq.Conf
	CoreRpc  zrpc.RpcClientConf
	Static   StaticConf
	Oss      OssConf
	// RedisConf 网关用 Redis 校验登录令牌（鉴权）
	RedisConf redisx.Conf
	// Elastic 聊天记录全文检索（可选，未配置则不启用）
	Elastic es.Conf
	// WsAllowedOrigins WebSocket 允许的 Origin 主机名白名单（为空时仅允许 localhost/127.0.0.1）
	WsAllowedOrigins []string
	// NodeID 网关节点标识：多网关水平扩展时用于路由表与 etcd 注册；为空则回退 hostname-port
	NodeID string `json:",optional"`
	// Registry etcd 网关节点注册配置（可选，未配置则不注册）
	Registry registry.Conf `json:",optional"`
}
