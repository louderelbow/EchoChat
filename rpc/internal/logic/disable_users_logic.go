package logic

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"echochat/pkg/constants"
	"echochat/pkg/enum"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"
)

type DisableUsersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDisableUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DisableUsersLogic {
	return &DisableUsersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DisableUsersLogic) DisableUsers(in *core.AbleUsersReq) (*core.BaseResp, error) {
	for _, uuid := range in.UuidList {
		if err := l.svcCtx.UserInfoModel.UpdateStatusByUuid(l.ctx, int64(enum.UserDisable), uuid); err != nil {
			l.Errorf(err.Error())
			return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
		}
		// 禁用用户时同时软删除其所有会话
		if err := l.svcCtx.SessionModel.UpdateDeletedAtBySendOrReceive(l.ctx, time.Now(), uuid); err != nil {
			l.Errorf(err.Error())
			return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
		}
	}
	return &core.BaseResp{Code: 200, Message: "禁用用户成功"}, nil
}
