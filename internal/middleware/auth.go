package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"

	"echochat/pkg/authtoken"
)

type ctxKey string

const (
	// CtxUuidKey 鉴权通过后写入 context 的调用者 uuid
	CtxUuidKey ctxKey = "auth_uuid"
	// CtxIsAdminKey 鉴权通过后写入 context 的调用者管理员标记
	CtxIsAdminKey ctxKey = "auth_is_admin"
)

// Enabled 是否开启鉴权：默认关闭，保证本地开发可直接运行；生产设 ECHO_AUTH=on 开启。
func Enabled() bool {
	return strings.EqualFold(os.Getenv("ECHO_AUTH"), "on")
}

// TokenFromRequest 从 Authorization 头或 ?token= 取令牌
func TokenFromRequest(r *http.Request) string {
	if h := r.Header.Get("Authorization"); h != "" {
		return strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
	}
	return r.URL.Query().Get("token")
}

// UuidFromCtx 获取经过鉴权的调用者 uuid（未开启鉴权时为空串，表示不校验）
func UuidFromCtx(ctx context.Context) string {
	if v, ok := ctx.Value(CtxUuidKey).(string); ok {
		return v
	}
	return ""
}

// IsOwner 校验调用者是否为指定 owner（如群主）：
// 未开启鉴权时直接放行（本地开发）；开启时必须令牌中的 uuid 与之完全一致。
func IsOwner(ctx context.Context, ownerId string) bool {
	if !Enabled() {
		return true
	}
	uuid := UuidFromCtx(ctx)
	return uuid != "" && uuid == ownerId
}

// 免鉴权的路径：登录/注册/验证码/WS 握手/退出登录/静态资源
var whitelist = map[string]bool{
	"/login":            true,
	"/register":         true,
	"/user/smsLogin":    true,
	"/user/sendSmsCode": true,
	"/user/wsLogout":    true,
	"/wss":              true,
	"/metrics":          true,
}

// 仅管理员可访问的路径（RBAC）
var adminOnly = map[string]bool{
	"/user/getUserInfoList":   true,
	"/user/disableUsers":      true,
	"/user/deleteUsers":       true,
	"/user/setAdmin":          true,
	"/user/ableUsers":         true,
	"/group/getGroupInfoList": true,
	"/group/deleteGroups":     true,
	"/group/setGroupsStatus":  true,
}

// Auth 返回 RBAC 鉴权中间件：未开启时直接放行；开启时校验令牌，并对管理端路径要求管理员身份。
func Auth() rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if !Enabled() {
				next(w, r)
				return
			}
			if r.Method == http.MethodOptions {
				next(w, r)
				return
			}
			path := r.URL.Path
			if whitelist[path] || strings.HasPrefix(path, "/static/") {
				next(w, r)
				return
			}

			uuid, isAdmin, ok := authtoken.Lookup(TokenFromRequest(r))
			if !ok {
				httpx.OkJsonCtx(r.Context(), w, map[string]interface{}{
					"code": 401, "message": "登录状态已失效，请重新登录",
				})
				return
			}
			if adminOnly[path] && !isAdmin {
				httpx.OkJsonCtx(r.Context(), w, map[string]interface{}{
					"code": 403, "message": "无权限操作",
				})
				return
			}
			ctx := context.WithValue(r.Context(), CtxUuidKey, uuid)
			ctx = context.WithValue(ctx, CtxIsAdminKey, isAdmin)
			next(w, r.WithContext(ctx))
		}
	}
}
