package canal

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"

	"echochat/pkg/redisx"
	"echochat/rpc/internal/svc"
)

// Start 启动 Canal(binlog) 变更事件消费者。
// Canal 中间件（MQ 模式，docker 部署）解析 MySQL binlog，以 JSON 事件投递到 RabbitMQ；
// 这里消费事件后，按「变更行数据」精准删除对应的 Redis 缓存 key。
// 缓存本身都带短 TTL（见 constants.REDIS_TIMEOUT），即使漏删也会自动过期，
// 因此本链路只负责「尽快失效」，把一致性兜底交给 TTL。
func Start(svcCtx *svc.ServiceContext) {
	queue := svcCtx.Config.RabbitMQ.CanalQueue
	if queue == "" {
		return
	}
	logx.Infof("Canal 缓存失效消费者已启动，队列：%s", queue)
	svcCtx.Mq.ConsumeCanal(queue, handleEvent)
}

// changeEvent 对应 Canal flatMessage(JSON) 事件结构
type changeEvent struct {
	Database string           `json:"database"`
	Table    string           `json:"table"`
	IsDdl    bool             `json:"isDdl"`
	Type     string           `json:"type"` // INSERT / UPDATE / DELETE
	Data     []map[string]any `json:"data"` // 变更后镜像（DELETE 时为被删行）
	Old      []map[string]any `json:"old"`  // 变更前镜像（UPDATE 才有）
}

func handleEvent(body []byte) error {
	var ev changeEvent
	if err := json.Unmarshal(body, &ev); err != nil {
		return err
	}
	// 只处理业务库的非 DDL 变更
	if ev.IsDdl || ev.Database != "echo_chat" {
		return nil
	}
	keys := affectedKeys(&ev)
	if len(keys) == 0 {
		return nil
	}
	if err := redisx.DelKeys(keys...); err != nil {
		return err
	}
	logx.Infof("Canal 事件 %s.%s(%s) 失效缓存 %d 个", ev.Database, ev.Table, ev.Type, len(keys))
	return nil
}

// affectedKeys 依据变更表与行数据，算出需要失效的缓存 key 集合。
// 关键点：UPDATE 的前镜像(old)同样参与计算，覆盖「成员被移除」「群主变更」等
// 只在前镜像里才存在受影响 id 的场景。
func affectedKeys(ev *changeEvent) []string {
	set := make(map[string]struct{})
	// addKey 只有 id 非空时才拼接 key。
	// 注意：UPDATE 事件的前镜像(old)只包含被修改的列，像 uuid/owner_id 这类列会缺失，
	// 若直接 "prefix"+"" 会得到空前缀垃圾 key（如 "group_info_"），因此必须判断 id。
	addKey := func(prefix, id string) {
		if id != "" {
			set[prefix+id] = struct{}{}
		}
	}
	add := func(key string) {
		if key != "" {
			set[key] = struct{}{}
		}
	}

	rows := make([]map[string]any, 0, len(ev.Data)+len(ev.Old))
	rows = append(rows, ev.Data...)
	rows = append(rows, ev.Old...)

	switch ev.Table {
	case "user_info":
		// 用户信息缓存
		for _, r := range rows {
			addKey("user_info_", str(r, "uuid"))
		}

	case "group_info":
		for _, r := range rows {
			// 群详情与群成员列表缓存
			addKey("group_info_", str(r, "uuid"))
			addKey("group_memberlist_", str(r, "uuid"))
			// 群主的「我拥有的群」列表
			addKey("contact_mygroup_list_", str(r, "owner_id"))
			// 群名称/头像/成员变化会影响所有成员的「我加入的群」列表
			for _, m := range membersOf(r) {
				addKey("my_joined_group_list_", m)
			}
		}

	case "user_contact":
		for _, r := range rows {
			userId := str(r, "user_id")
			contactId := str(r, "contact_id")
			// 双向：变更行的 owner 与对方都可能受影响
			addKey("contact_user_list_", userId)
			addKey("contact_user_list_", contactId)
			// 群联系（contact_id 以 G 开头）变化会影响该用户「我加入的群」列表
			if strings.HasPrefix(contactId, "G") {
				addKey("my_joined_group_list_", userId)
			}
		}

	case "session":
		for _, r := range rows {
			sendId := str(r, "send_id")
			receiveId := str(r, "receive_id")
			// 会话列表缓存（单聊/群聊两份列表）
			addKey("session_list_", sendId)
			if strings.HasPrefix(receiveId, "G") {
				addKey("group_session_list_", sendId)
			}
			// 单个会话详情缓存，正反两个方向都失效
			if sendId != "" && receiveId != "" {
				add("session_" + sendId + "_" + receiveId)
				add("session_" + receiveId + "_" + sendId)
			}
		}
	}

	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	return keys
}

// str 读取列值并转字符串（Canal flatMessage 的列值可能是字符串或 JSON 原生类型）
func str(row map[string]any, col string) string {
	v, ok := row[col]
	if !ok || v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
}

// membersOf 解析 group_info.members（JSON 字符串数组）
func membersOf(row map[string]any) []string {
	raw := str(row, "members")
	if raw == "" {
		return nil
	}
	var list []string
	if err := json.Unmarshal([]byte(raw), &list); err != nil {
		return nil
	}
	return list
}
