package chat

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"echochat/internal/model"
	"echochat/pkg/constants"
	"echochat/pkg/dto"
	"echochat/pkg/enum"
	"echochat/pkg/mq"
	"echochat/pkg/random"
	"echochat/pkg/redisx"
	"echochat/pkg/route"
	"echochat/rpc/internal/svc"
)

// Start 启动上行消息消费者：持久化 -> 写 Redis 缓存 -> 定向投递到网关节点
func Start(svcCtx *svc.ServiceContext) error {
	return svcCtx.Mq.ConsumeChat(func(body []byte) error {
		return handle(svcCtx, body)
	})
}

func normalizePath(path string) string {
	if path == "https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png" {
		return path
	}
	if i := strings.Index(path, "/static/"); i >= 0 {
		return path[i:]
	}
	return path
}

func handle(svcCtx *svc.ServiceContext, body []byte) error {
	var req dto.ChatMessageRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return err
	}
	switch req.Type {
	case enum.MsgText:
		return handleText(svcCtx, &req)
	case enum.MsgFile:
		return handleFile(svcCtx, &req)
	case enum.MsgAudioOrVideo:
		return handleAV(svcCtx, &req)
	}
	return nil
}

// normalizeTarget 上行消息容错：客户端漏传/传错 receive_id 时，用会话记录补全目标。
// session_id 对应的会话记录了 send_id/receive_id，服务端以它为准，避免消息落库后
// 无法路由、也进不了会话历史（不能假定客户端一定带上正确的 receive_id）。
func normalizeTarget(svcCtx *svc.ServiceContext, req *dto.ChatMessageRequest) {
	if strings.HasPrefix(req.ReceiveId, "U") || strings.HasPrefix(req.ReceiveId, "G") {
		return
	}
	s, err := svcCtx.SessionModel.FindOneNotDeletedByUuid(context.Background(), req.SessionId)
	if err != nil {
		logx.Errorf("消息接收方缺失且补全失败(session_id=%s): %v", req.SessionId, err)
		return
	}
	if req.SendId == "" {
		req.SendId = s.SendId
	}
	logx.Infof("消息接收方缺失，已按会话 %s 补全接收方为 %s", req.SessionId, s.ReceiveId)
	req.ReceiveId = s.ReceiveId
}

func handleText(svcCtx *svc.ServiceContext, req *dto.ChatMessageRequest) error {
	normalizeTarget(svcCtx, req)
	now := time.Now()
	msg := &model.Message{
		Uuid:       "M" + random.GetNowAndLenRandomString(11),
		SessionId:  req.SessionId,
		Type:       int64(enum.MsgText),
		Content:    sql.NullString{String: req.Content, Valid: true},
		SendId:     req.SendId,
		SendName:   req.SendName,
		SendAvatar: normalizePath(req.SendAvatar),
		ReceiveId:  req.ReceiveId,
		FileSize:   sql.NullString{String: "0B", Valid: true},
		Status:     int64(enum.MsgUnsent),
		CreatedAt:  now,
	}
	if _, err := svcCtx.MessageModel.Insert(context.Background(), msg); err != nil {
		return err
	}
	createdAt := now.Format("2006-01-02 15:04:05")
	// 异步同步到 ES，供聊天记录全文检索（失败不影响消息主流程）
	go indexMessageToES(svcCtx, msg, req.Content)

	if strings.HasPrefix(msg.ReceiveId, "U") {
		pushSingle(svcCtx, msg, dto.GetMessageListRespond{
			SendId: msg.SendId, SendName: msg.SendName, SendAvatar: req.SendAvatar,
			ReceiveId: msg.ReceiveId, Type: enum.MsgText, Content: req.Content,
			CreatedAt: createdAt,
		})
		return nil
	}
	if strings.HasPrefix(msg.ReceiveId, "G") {
		pushGroup(svcCtx, msg, dto.GetGroupMessageListRespond{
			SendId: msg.SendId, SendName: msg.SendName, SendAvatar: req.SendAvatar,
			ReceiveId: msg.ReceiveId, Type: enum.MsgText, Content: req.Content,
			CreatedAt: createdAt,
		})
		return nil
	}
	// 接收方非法：只落库不投递。这里绝不能返回 error——否则会被重新入队，
	// 造成「无限重投递 + 每轮又插入一条重复消息」（曾导致 3000+ 重复记录）。
	logx.Errorf("消息 %s 接收方非法(receive_id=%q)，仅落库不投递", msg.Uuid, msg.ReceiveId)
	return nil
}

func handleFile(svcCtx *svc.ServiceContext, req *dto.ChatMessageRequest) error {
	normalizeTarget(svcCtx, req)
	now := time.Now()
	msg := &model.Message{
		Uuid:       "M" + random.GetNowAndLenRandomString(11),
		SessionId:  req.SessionId,
		Type:       int64(enum.MsgFile),
		Content:    sql.NullString{String: "", Valid: true},
		Url:        sql.NullString{String: req.Url, Valid: true},
		SendId:     req.SendId,
		SendName:   req.SendName,
		SendAvatar: normalizePath(req.SendAvatar),
		ReceiveId:  req.ReceiveId,
		FileSize:   sql.NullString{String: req.FileSize, Valid: true},
		FileType:   sql.NullString{String: req.FileType, Valid: true},
		FileName:   sql.NullString{String: req.FileName, Valid: true},
		Status:     int64(enum.MsgUnsent),
		CreatedAt:  now,
	}
	if _, err := svcCtx.MessageModel.Insert(context.Background(), msg); err != nil {
		return err
	}
	createdAt := now.Format("2006-01-02 15:04:05")
	// 文件消息以文件名为检索内容同步到 ES
	go indexMessageToES(svcCtx, msg, req.FileName)

	if strings.HasPrefix(msg.ReceiveId, "U") {
		pushSingle(svcCtx, msg, dto.GetMessageListRespond{
			SendId: msg.SendId, SendName: msg.SendName, SendAvatar: req.SendAvatar,
			ReceiveId: msg.ReceiveId, Type: enum.MsgFile, Url: req.Url,
			FileSize: req.FileSize, FileType: req.FileType, FileName: req.FileName,
			CreatedAt: createdAt,
		})
		return nil
	}
	if strings.HasPrefix(msg.ReceiveId, "G") {
		pushGroup(svcCtx, msg, dto.GetGroupMessageListRespond{
			SendId: msg.SendId, SendName: msg.SendName, SendAvatar: req.SendAvatar,
			ReceiveId: msg.ReceiveId, Type: enum.MsgFile, Url: req.Url,
			FileSize: req.FileSize, FileType: req.FileType, FileName: req.FileName,
			CreatedAt: createdAt,
		})
		return nil
	}
	logx.Errorf("文件消息 %s 接收方非法(receive_id=%q)，仅落库不投递", msg.Uuid, msg.ReceiveId)
	return nil
}

// indexMessageToES 把消息异步写入 ES 供全文检索；失败仅记日志，不影响消息主流程
func indexMessageToES(svcCtx *svc.ServiceContext, msg *model.Message, content string) {
	if !svcCtx.Es.Enabled() {
		return
	}
	doc := map[string]any{
		"uuid":       msg.Uuid,
		"session_id": msg.SessionId,
		"receive_id": msg.ReceiveId,
		"send_id":    msg.SendId,
		"send_name":  msg.SendName,
		"content":    content,
		"created_at": msg.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if err := svcCtx.Es.IndexMessage(context.Background(), doc); err != nil {
		logx.Errorf("同步消息到 ES 失败: %v", err)
	}
}

func handleAV(svcCtx *svc.ServiceContext, req *dto.ChatMessageRequest) error {
	var avData dto.AVData
	if err := json.Unmarshal([]byte(req.AVdata), &avData); err != nil {
		return err
	}
	now := time.Now()
	msg := &model.Message{
		Uuid:       "M" + random.GetNowAndLenRandomString(11),
		SessionId:  req.SessionId,
		Type:       int64(enum.MsgAudioOrVideo),
		SendId:     req.SendId,
		SendName:   req.SendName,
		SendAvatar: normalizePath(req.SendAvatar),
		ReceiveId:  req.ReceiveId,
		Status:     int64(enum.MsgUnsent),
		CreatedAt:  now,
		AvData:     sql.NullString{String: req.AVdata, Valid: true},
	}
	// 与原实现一致：只有 PROXY 的 start/receive/reject 信令才落库
	if avData.MessageId == "PROXY" &&
		(avData.Type == "start_call" || avData.Type == "receive_call" || avData.Type == "reject_call") {
		if _, err := svcCtx.MessageModel.Insert(context.Background(), msg); err != nil {
			return err
		}
	}
	rsp := dto.AVMessageRespond{
		SendId: msg.SendId, SendName: msg.SendName, SendAvatar: msg.SendAvatar,
		ReceiveId: msg.ReceiveId, Type: enum.MsgAudioOrVideo,
		CreatedAt: now.Format("2006-01-02 15:04:05"), AVdata: req.AVdata,
	}
	payload, err := json.Marshal(rsp)
	if err != nil {
		return err
	}
	// 群聊通话信令：扇出给全体群成员（排除发起方，避免自己收到自己的来电）。
	// 之前这里只处理单聊（receive_id 以 U 开头），群聊信令被直接丢弃，
	// 导致一端在群里发起通话时其他成员完全收不到来电提示。
	if strings.HasPrefix(msg.ReceiveId, "G") {
		deliverGroupAV(svcCtx, msg, payload)
		return nil
	}
	if !strings.HasPrefix(msg.ReceiveId, "U") {
		return nil
	}
	// 单聊通话信令不回显给发送方，否则会出现两个 start_call
	deliver(svcCtx, msg.ReceiveId, payload)
	return nil
}

// deliverGroupAV 群聊通话信令投递：查群成员后逐个投递，排除发起方。
// 与消息投递一致，任何失败都只记日志、不返回错误，避免触发重投递死循环。
func deliverGroupAV(svcCtx *svc.ServiceContext, msg *model.Message, payload []byte) {
	group, err := svcCtx.GroupInfoModel.FindOneMembersByUuid(context.Background(), msg.ReceiveId)
	if err != nil {
		logx.Errorf("群通话信令 %s 查群失败(receive_id=%s): %v", msg.Uuid, msg.ReceiveId, err)
		return
	}
	var members []string
	if group.Members.Valid {
		_ = json.Unmarshal([]byte(group.Members.String), &members)
	}
	targets := make([]string, 0, len(members))
	for _, m := range members {
		if m != "" && m != msg.SendId {
			targets = append(targets, m)
		}
	}
	deliverBatch(svcCtx, targets, payload)
}

// deliver 按「用户->节点」路由表把下行信封投递到目标网关节点。
// 用户不在线（路由缺失/过期）时直接跳过：消息已落库，登录后由历史消息拉取兜底。
// 注意：这里绝不返回错误，否则 ConsumeChat 会 Nack 重入队导致死循环。
func deliver(svcCtx *svc.ServiceContext, uuid string, payload []byte) {
	if uuid == "" {
		return
	}
	node, err := route.Lookup(uuid)
	if err != nil {
		logx.Errorf("查询用户路由失败(uuid=%s): %v", uuid, err)
		return
	}
	if node == "" {
		logx.Infof("用户 %s 不在线，跳过下行投递", uuid)
		return
	}
	env, err := json.Marshal(mq.PushEnvelope{To: uuid, Body: payload})
	if err != nil {
		logx.Errorf("封装下行信封失败(uuid=%s): %v", uuid, err)
		return
	}
	if err := svcCtx.Mq.PublishPushTo(node, env); err != nil {
		logx.Errorf("投递到网关节点 %s 失败: %v", node, err)
	}
}

// deliverBatch 群聊批量投递：先用 MGET 一次性查所有成员的路由，再逐个投递。
func deliverBatch(svcCtx *svc.ServiceContext, uuids []string, payload []byte) {
	routes, err := route.BatchLookup(uuids)
	if err != nil {
		logx.Errorf("批量查询用户路由失败: %v", err)
		return
	}
	for uuid, node := range routes {
		env, err := json.Marshal(mq.PushEnvelope{To: uuid, Body: payload})
		if err != nil {
			logx.Errorf("封装下行信封失败(uuid=%s): %v", uuid, err)
			continue
		}
		if err := svcCtx.Mq.PublishPushTo(node, env); err != nil {
			logx.Errorf("投递到网关节点 %s 失败: %v", node, err)
		}
	}
}

// pushSingle 单聊下行：投递给接收方，发送方在线则回显。
// 注意：消息此时已落库，投递阶段的任何失败都只记日志、绝不向上返回错误，
// 否则消费失败会被 RabbitMQ 重新入队 → 无限重投递 + 每轮重复落库。
func pushSingle(svcCtx *svc.ServiceContext, msg *model.Message, rsp dto.GetMessageListRespond) {
	payload, err := json.Marshal(rsp)
	if err != nil {
		logx.Errorf("序列化单聊下行消息失败(uuid=%s): %v", msg.Uuid, err)
		return
	}
	// 跨节点投递给接收方；发送方在线则回显
	deliver(svcCtx, msg.ReceiveId, payload)
	deliver(svcCtx, msg.SendId, payload)
	appendMessageCache("message_list_"+msg.SendId+"_"+msg.ReceiveId, rsp)
}

// pushGroup 群聊下行：逐个成员投递。查群失败（如群已解散/传入非法群 id）同样只记日志、不返回错误。
func pushGroup(svcCtx *svc.ServiceContext, msg *model.Message, rsp dto.GetGroupMessageListRespond) {
	payload, err := json.Marshal(rsp)
	if err != nil {
		logx.Errorf("序列化群聊下行消息失败(uuid=%s): %v", msg.Uuid, err)
		return
	}
	group, err := svcCtx.GroupInfoModel.FindOneByUuid(context.Background(), msg.ReceiveId)
	if err != nil {
		logx.Errorf("群消息 %s 查群失败(receive_id=%s): %v", msg.Uuid, msg.ReceiveId, err)
		return
	}
	var members []string
	if group.Members.Valid {
		_ = json.Unmarshal([]byte(group.Members.String), &members)
	}
	deliverBatch(svcCtx, members, payload)
	appendGroupMessageCache("group_messagelist_"+msg.ReceiveId, rsp)
}

func appendMessageCache(key string, item dto.GetMessageListRespond) {
	s, err := redisx.GetKeyNilIsErr(key)
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			logx.Errorf("redis get %s error: %v", key, err)
		}
		return
	}
	var list []dto.GetMessageListRespond
	if err := json.Unmarshal([]byte(s), &list); err != nil {
		logx.Errorf("redis unmarshal %s error: %v", key, err)
		return
	}
	list = append(list, item)
	b, _ := json.Marshal(list)
	if err := redisx.SetKeyEx(key, string(b), time.Minute*constants.REDIS_TIMEOUT); err != nil {
		logx.Errorf("redis set %s error: %v", key, err)
	}
}

func appendGroupMessageCache(key string, item dto.GetGroupMessageListRespond) {
	s, err := redisx.GetKeyNilIsErr(key)
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			logx.Errorf("redis get %s error: %v", key, err)
		}
		return
	}
	var list []dto.GetGroupMessageListRespond
	if err := json.Unmarshal([]byte(s), &list); err != nil {
		logx.Errorf("redis unmarshal %s error: %v", key, err)
		return
	}
	list = append(list, item)
	b, _ := json.Marshal(list)
	if err := redisx.SetKeyEx(key, string(b), time.Minute*constants.REDIS_TIMEOUT); err != nil {
		logx.Errorf("redis set %s error: %v", key, err)
	}
}
