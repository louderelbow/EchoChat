package password

import (
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// Hash 生成密码的 bcrypt 哈希（用于注册/修改密码时存储）
func Hash(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Verify 校验明文密码与存储的哈希是否匹配
func Verify(hashed, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)) == nil
}

// IsHashed 判断存储值是否已是 bcrypt 哈希（用于兼容历史明文密码）
func IsHashed(s string) bool {
	return len(s) == 60 && strings.HasPrefix(s, "$2")
}
