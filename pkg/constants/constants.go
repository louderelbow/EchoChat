package constants

import "time"

const (
	CHANNEL_SIZE  = 100            // 通道大小
	SYSTEM_ERROR  = "系统错误，请联系工作人员" // 系统错误
	FILE_MAX_SIZE = 50 * 1024 * 1024 // 上传文件最大大小（字节），50MB，需与前端限制一致
	REDIS_TIMEOUT = 1              // redis 消息缓存过期（分钟）

	// WebSocket 心跳：服务端按 WS_PING_PERIOD 发协议级 Ping，浏览器自动回 Pong；
	// 超过 WS_PONG_WAIT 没收到任何数据即判定为死连接并清理（无需前端改动）。
	WS_PING_PERIOD = 25 * time.Second
	WS_PONG_WAIT   = 60 * time.Second

	// ROUTE_RENEW_INTERVAL 网关续期「用户->节点」路由表的间隔（须显著小于 route.RouteTTL）
	ROUTE_RENEW_INTERVAL = 20 * time.Second

	// DEFAULT_AVATAR 默认头像 URL（与 schema/echo.sql 中 user_info、group_info 的默认值一致）。
	// 群头像为空时必须回退到它：前端拿到空 avatar 会因缺字段而报错，导致整个列表渲染失败。
	DEFAULT_AVATAR = "https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png"
)
