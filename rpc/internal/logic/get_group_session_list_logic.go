package logic

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"echochat/pkg/constants"
	"echochat/pkg/redisx"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetGroupSessionListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetGroupSessionListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupSessionListLogic {
	return &GetGroupSessionListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetGroupSessionListLogic) GetGroupSessionList(in *core.OwnlistReq) (*core.GetGroupSessionListResp, error) {
	ownerId := in.OwnerId
	cacheKey := "group_session_list_" + ownerId
	rspString, err := redisx.GetKeyNilIsErr(cacheKey)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			sessionList, qerr := l.svcCtx.SessionModel.FindBySendId(l.ctx, ownerId)
			if qerr != nil {
				l.Error(qerr.Error())
				return &core.GetGroupSessionListResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
			}
			sessions := make([]*core.GroupSessionItem, 0, len(sessionList))
			for _, session := range sessionList {
				if len(session.ReceiveId) > 0 && session.ReceiveId[0] == 'G' {
					sessions = append(sessions, &core.GroupSessionItem{
						SessionId: session.Uuid,
						Avatar:    defaultAvatarIfEmpty(session.Avatar),
						GroupId:   session.ReceiveId,
						GroupName: session.ReceiveName,
					})
				}
			}
			if data, merr := json.Marshal(sessions); merr != nil {
				l.Error(merr.Error())
			} else if serr := redisx.SetKeyEx(cacheKey, string(data), time.Minute*constants.REDIS_TIMEOUT); serr != nil {
				l.Error(serr.Error())
			}
			return &core.GetGroupSessionListResp{Code: 200, Message: "获取成功", Sessions: sessions}, nil
		}
		l.Error(err.Error())
		return &core.GetGroupSessionListResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	var sessions []*core.GroupSessionItem
	if err := json.Unmarshal([]byte(rspString), &sessions); err != nil {
		l.Error(err.Error())
	}
	return &core.GetGroupSessionListResp{Code: 200, Message: "获取成功", Sessions: sessions}, nil
}
