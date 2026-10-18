package authtoken

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"echochat/pkg/redisx"
)

// KeyPrefix 登录令牌在 Redis 中的键前缀
const KeyPrefix = "login_token_"

// TTL 令牌有效期
const TTL = 24 * time.Hour

// Issue 生成登录令牌并写入 Redis，值格式为 "uuid|isAdmin"，供网关按 RBAC 校验。
func Issue(uuid string, isAdmin bool) (string, error) {
	token, err := newToken()
	if err != nil {
		return "", err
	}
	val := uuid + "|" + strconv.FormatBool(isAdmin)
	if err := redisx.SetKeyEx(KeyPrefix+token, val, TTL); err != nil {
		return "", err
	}
	return token, nil
}

// Lookup 校验令牌，返回 uuid、是否管理员；ok=false 表示令牌无效。
func Lookup(token string) (uuid string, isAdmin bool, ok bool) {
	if token == "" {
		return "", false, false
	}
	v, err := redisx.GetKey(KeyPrefix + token)
	if err != nil || v == "" {
		return "", false, false
	}
	parts := strings.SplitN(v, "|", 2)
	uuid = parts[0]
	if len(parts) == 2 {
		isAdmin = parts[1] == "true"
	}
	return uuid, isAdmin, uuid != ""
}

// Revoke 注销令牌（退出登录时调用）。
func Revoke(token string) error {
	if token == "" {
		return nil
	}
	return redisx.DelKeyIfExists(KeyPrefix + token)
}

func newToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
