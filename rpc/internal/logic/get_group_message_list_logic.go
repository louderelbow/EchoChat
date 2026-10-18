package logic

import (
	"context"
	"encoding/json"
	"errors"

	"echochat/pkg/constants"
	"echochat/pkg/redisx"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetGroupMessageListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetGroupMessageListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupMessageListLogic {
	return &GetGroupMessageListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetGroupMessageListLogic) GetGroupMessageList(in *core.GetGroupMessageListReq) (*core.GetMessageListResp, error) {
	groupId := in.GroupId
	cacheKey := "group_messagelist_" + groupId
	rspString, err := redisx.GetKeyNilIsErr(cacheKey)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			messageList, qerr := l.svcCtx.MessageModel.FindByReceiveId(l.ctx, groupId)
			if qerr != nil {
				l.Error(qerr.Error())
				return &core.GetMessageListResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
			}
			messages := make([]*core.MessageItem, 0, len(messageList))
			for _, message := range messageList {
				messages = append(messages, &core.MessageItem{
					SendId:     message.SendId,
					SendName:   message.SendName,
					SendAvatar: defaultAvatarIfEmpty(message.SendAvatar),
					ReceiveId:  message.ReceiveId,
					Content:    message.Content.String,
					Url:        message.Url.String,
					Type:       int32(message.Type),
					FileType:   message.FileType.String,
					FileName:   message.FileName.String,
					FileSize:   message.FileSize.String,
					CreatedAt:  message.CreatedAt.Format("2006-01-02 15:04:05"),
				})
			}
			return &core.GetMessageListResp{Code: 200, Message: "获取聊天记录成功", Messages: messages}, nil
		}
		l.Error(err.Error())
		return &core.GetMessageListResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	var messages []*core.MessageItem
	if err := json.Unmarshal([]byte(rspString), &messages); err != nil {
		l.Error(err.Error())
	}
	return &core.GetMessageListResp{Code: 200, Message: "获取聊天记录成功", Messages: messages}, nil
}
