package route

import (
	"time"

	"echochat/pkg/redisx"
)

// 用户 -> 网关节点 路由表。
// key: im:route:u:<uuid>   value: nodeID
// TTL 即「在线」信号：网关周期性续期，超时未续期视为离线，避免死连接把消息投给幽灵节点。
const (
	keyPrefix = "im:route:u:"
	// RouteTTL 路由条目存活时间：网关每 RouteTTL/3 左右续期一次
	RouteTTL = 60 * time.Second
)

func key(uuid string) string {
	return keyPrefix + uuid
}

// Set 记录用户当前所在网关节点（后连上者覆盖，last-writer-wins）
func Set(uuid, nodeID string, ttl time.Duration) error {
	if uuid == "" || nodeID == "" {
		return nil
	}
	return redisx.SetKeyEx(key(uuid), nodeID, ttl)
}

// Lookup 查询用户所在节点；不存在或已过期返回空串（视为离线）
func Lookup(uuid string) (string, error) {
	return redisx.GetKey(key(uuid))
}

// BatchLookup 批量查询（群聊扇出用），只返回命中的 uuid -> nodeID
func BatchLookup(uuids []string) (map[string]string, error) {
	result := make(map[string]string, len(uuids))
	if len(uuids) == 0 {
		return result, nil
	}
	keys := make([]string, len(uuids))
	for i, u := range uuids {
		keys[i] = key(u)
	}
	vals, err := redisx.MGet(keys...)
	if err != nil {
		return nil, err
	}
	for i, u := range uuids {
		if i < len(vals) && vals[i] != "" {
			result[u] = vals[i]
		}
	}
	return result, nil
}

// Renew 批量续期本节点上的路由（心跳刷新），返回首个错误但不中断
func Renew(uuids []string, nodeID string, ttl time.Duration) error {
	var firstErr error
	for _, u := range uuids {
		if err := Set(u, nodeID, ttl); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// CompareAndDel 仅当路由仍指向 expectedNode 时才删除，返回是否真的删除。
// 跨节点场景：用户已连到新节点时，旧节点的下线清理不会误删新路由。
func CompareAndDel(uuid, expectedNode string) (bool, error) {
	return redisx.CompareAndDel(key(uuid), expectedNode)
}

// DeleteByNode 删除所有指向某节点的路由（节点故障清理），返回删除数量。
// 由 etcd watch 在节点 key 消失时触发，把故障恢复从「等 TTL」降到「秒级」。
func DeleteByNode(nodeID string) (int, error) {
	if nodeID == "" {
		return 0, nil
	}
	keys, err := redisx.ScanKeys(keyPrefix + "*")
	if err != nil {
		return 0, err
	}
	deleted := 0
	for _, k := range keys {
		ok, err := redisx.CompareAndDel(k, nodeID)
		if err != nil {
			return deleted, err
		}
		if ok {
			deleted++
		}
	}
	return deleted, nil
}
