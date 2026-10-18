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

type GetUserSessionListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserSessionListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserSessionListLogic {
	return &GetUserSessionListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserSessionListLogic) GetUserSessionList(in *core.OwnlistReq) (*core.GetUserSessionListResp, error) {
	ownerId := in.OwnerId
	cacheKey := "session_list_" + ownerId
	rspString, err := redisx.GetKeyNilIsErr(cacheKey)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			sessionList, qerr := l.svcCtx.SessionModel.FindBySendId(l.ctx, ownerId)
			if qerr != nil {
				l.Error(qerr.Error())
				return &core.GetUserSessionListResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
			}
			sessions := make([]*core.UserSessionItem, 0, len(sessionList))
			for _, session := range sessionList {
				if len(session.ReceiveId) > 0 && session.ReceiveId[0] == 'U' {
					sessions = append(sessions, &core.UserSessionItem{
						SessionId: session.Uuid,
						Avatar:    session.Avatar,
						UserId:    session.ReceiveId,
						UserName:  session.ReceiveName,
					})
				}
			}
			if data, merr := json.Marshal(sessions); merr != nil {
				l.Error(merr.Error())
			} else if serr := redisx.SetKeyEx(cacheKey, string(data), time.Minute*constants.REDIS_TIMEOUT); serr != nil {
				l.Error(serr.Error())
			}
			return &core.GetUserSessionListResp{Code: 200, Message: "获取成功", Sessions: sessions}, nil
		}
		l.Error(err.Error())
		return &core.GetUserSessionListResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	var sessions []*core.UserSessionItem
	if err := json.Unmarshal([]byte(rspString), &sessions); err != nil {
		l.Error(err.Error())
	}
	return &core.GetUserSessionListResp{Code: 200, Message: "获取成功", Sessions: sessions}, nil
}
