// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"echochat/internal/config"
	"echochat/internal/handler"
	"echochat/internal/middleware"
	"echochat/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/prometheus"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

var configFile = flag.String("f", "etc/echo-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	// UseEnv 使配置支持 ${VAR} 环境变量替换（OSS 的 AK/SK 即通过 ACID/ACSC 注入）
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	// 生产可设 ECHO_TLS=off 关闭网关自带 TLS，改由前置 nginx 终结 TLS（网关走明文 HTTP）
	if strings.EqualFold(os.Getenv("ECHO_TLS"), "off") {
		c.CertFile = ""
		c.KeyFile = ""
	}

	// 统一错误响应格式，保持与原 Gin 版 JsonBack 一致：HTTP 恒 200 + {code,message}
	// 注意：这里必须把底层错误打进日志，否则所有异常都会被统一成「系统错误」而无法排查
	httpx.SetErrorHandler(func(err error) (int, interface{}) {
		logx.Errorf("请求处理失败: %v", err)
		return http.StatusOK, map[string]interface{}{
			"code":    500,
			"message": "系统错误，请联系工作人员",
		}
	})

	// 启用 go-zero 内置 Prometheus 采集（连接数/下行推送等 IM 业务指标），并通过 /metrics 暴露
	prometheus.Enable()

	// 开启 CORS（含 OPTIONS 预检），供前端跨域调用本网关
	server := rest.MustNewServer(c.RestConf, rest.WithCors())
	defer server.Stop()
	// RBAC 鉴权中间件：默认关闭（ECHO_AUTH != on 时放行），生产开启后校验令牌与管理员权限
	server.Use(middleware.Auth())

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)
	registerExtraRoutes(server, ctx, c)

	// 启动跨节点投递消费者：把本节点队列收到的「信封」拆壳后推给本地在线连接
	if err := ctx.Mq.ConsumePush(ctx.Hub.OnPush); err != nil {
		panic(err)
	}

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
