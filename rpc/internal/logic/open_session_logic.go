package logic

import (
	"context"
	"encoding/json"
	"errors"

	"echochat/internal/model"
	"echochat/pkg/constants"
	"echochat/pkg/enum"
	"echochat/pkg/random"
	"echochat/pkg/redisx"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type OpenSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOpenSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OpenSessionLogic {
	return &OpenSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 会话
func (l *OpenSessionLogic) OpenSession(in *core.OpenSessionReq) (*core.OpenSessionResp, error) {
	rspString, err := redisx.GetKeyNilIsErr("session_" + in.SendId + "_" + in.ReceiveId)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			session, qerr := l.svcCtx.SessionModel.FindOneBySendReceive(l.ctx, in.SendId, in.ReceiveId)
			if qerr != nil {
				if errors.Is(qerr, sqlx.ErrNotFound) {
					// 会话没有找到，将新建会话
					_, uerr := l.svcCtx.UserInfoModel.FindOneAllByUuid(l.ctx, in.SendId)
					if uerr != nil {
						l.Error(uerr.Error())
						return &core.OpenSessionResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
					}
					newSession := &model.Session{
						Uuid:      "S" + random.GetNowAndLenRandomString(11),
						SendId:    in.SendId,
						ReceiveId: in.ReceiveId,
					}
					if len(in.ReceiveId) > 0 && in.ReceiveId[0] == 'U' {
						receiveUser, rerr := l.svcCtx.UserInfoModel.FindOneAllByUuid(l.ctx, in.ReceiveId)
						if rerr != nil {
							l.Error(rerr.Error())
							return &core.OpenSessionResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
						}
						if receiveUser.Status == int64(enum.UserDisable) {
							l.Error("该用户被禁用了")
							return &core.OpenSessionResp{Code: 400, Message: "该用户被禁用了"}, nil
						}
						newSession.ReceiveName = receiveUser.Nickname
						newSession.Avatar = receiveUser.Avatar
					} else {
						receiveGroup, rerr := l.svcCtx.GroupInfoModel.FindOneAllByUuid(l.ctx, in.ReceiveId)
						if rerr != nil {
							l.Error(rerr.Error())
							return &core.OpenSessionResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
						}
						if receiveGroup.Status == int64(enum.GroupDisable) {
							l.Error("该群聊被禁用了")
							return &core.OpenSessionResp{Code: 400, Message: "该群聊被禁用了"}, nil
						}
						newSession.ReceiveName = receiveGroup.Name
						newSession.Avatar = receiveGroup.Avatar
					}
					if _, ierr := l.svcCtx.SessionModel.Insert(l.ctx, newSession); ierr != nil {
						l.Error(ierr.Error())
						return &core.OpenSessionResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
					}
					return &core.OpenSessionResp{Code: 200, Message: "会话创建成功", SessionId: newSession.Uuid}, nil
				}
				// 与原实现一致：非“未找到”的查询异常直接返回（此时会话为空）
				return &core.OpenSessionResp{Code: 200, Message: "会话创建成功", SessionId: session.Uuid}, nil
			}
			return &core.OpenSessionResp{Code: 200, Message: "会话创建成功", SessionId: session.Uuid}, nil
		}
		l.Error(err.Error())
		return &core.OpenSessionResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	var session model.Session
	if err := json.Unmarshal([]byte(rspString), &session); err != nil {
		l.Error(err.Error())
	}
	return &core.OpenSessionResp{Code: 200, Message: "会话创建成功", SessionId: session.Uuid}, nil
}
