package redisx

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Conf struct {
	Host     string
	Port     int
	Password string
	Db       int
}

var client *redis.Client
var ctx = context.Background()

func Init(conf Conf) {
	client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", conf.Host, conf.Port),
		Password: conf.Password,
		DB:       conf.Db,
	})
}

func SetKeyEx(key string, value string, timeout time.Duration) error {
	return client.Set(ctx, key, value, timeout).Err()
}

// GetKey 不存在时返回空串与 nil（与原实现一致）
func GetKey(key string) (string, error) {
	value, err := client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", nil
		}
		return "", err
	}
	return value, nil
}

// GetKeyNilIsErr 不存在时返回 redis.Nil
func GetKeyNilIsErr(key string) (string, error) {
	return client.Get(ctx, key).Result()
}

func DelKeyIfExists(key string) error {
	exists, err := client.Exists(ctx, key).Result()
	if err != nil {
		return err
	}
	if exists == 1 {
		return client.Del(ctx, key).Err()
	}
	return nil
}

func DelKeysWithPrefix(prefix string) error {
	keys, err := client.Keys(ctx, prefix+"*").Result()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return client.Del(ctx, keys...).Err()
	}
	return nil
}

// DelKeys 批量删除指定 key，用于 binlog 变更后的精准缓存失效
func DelKeys(keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return client.Del(ctx, keys...).Err()
}

// MGet 批量读取：不存在的 key 返回空串（长度与 keys 一致）
func MGet(keys ...string) ([]string, error) {
	if len(keys) == 0 {
		return nil, nil
	}
	vals, err := client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	out := make([]string, len(vals))
	for i, v := range vals {
		if s, ok := v.(string); ok {
			out[i] = s
		}
	}
	return out, nil
}

// ScanKeys 用 SCAN 游标遍历匹配 pattern 的 key（替代阻塞的 KEYS）
func ScanKeys(pattern string) ([]string, error) {
	var keys []string
	var cursor uint64
	for {
		batch, next, err := client.Scan(ctx, cursor, pattern, 200).Result()
		if err != nil {
			return nil, err
		}
		keys = append(keys, batch...)
		cursor = next
		if cursor == 0 {
			return keys, nil
		}
	}
}

// compareAndDelScript 原子「比较并删除」：仅当 key 当前值等于期望值时才删除
var compareAndDelScript = redis.NewScript(`
if redis.call('GET', KEYS[1]) == ARGV[1] then
  return redis.call('DEL', KEYS[1])
end
return 0
`)

// CompareAndDel 仅当 key 的当前值 == expected 时才删除，返回是否真的删除。
// 用于路由表下线清理：避免旧节点误删已被新节点抢占的路由。
func CompareAndDel(key, expected string) (bool, error) {
	n, err := compareAndDelScript.Run(ctx, client, []string{key}, expected).Int()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}
	return n > 0, nil
}
