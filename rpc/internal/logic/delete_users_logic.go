package logic

import (
	"context"
	"time"

	"echochat/pkg/constants"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteUsersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteUsersLogic {
	return &DeleteUsersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteUsers 删除用户：软删除用户本身，并级联软删除其会话、联系人、申请记录
func (l *DeleteUsersLogic) DeleteUsers(in *core.AbleUsersReq) (*core.BaseResp, error) {
	now := time.Now()
	for _, uuid := range in.UuidList {
		if err := l.svcCtx.UserInfoModel.SoftDeleteByUuid(l.ctx, now, uuid); err != nil {
			logx.Errorf("delete user %s error: %v", uuid, err)
			return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
		}
		if err := l.svcCtx.SessionModel.UpdateDeletedAtBySendOrReceive(l.ctx, now, uuid); err != nil {
			logx.Errorf("delete user %s sessions error: %v", uuid, err)
			return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
		}
		if err := l.svcCtx.UserContactModel.SoftDeleteByUserOrContact(l.ctx, now, uuid, uuid); err != nil {
			logx.Errorf("delete user %s contacts error: %v", uuid, err)
			return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
		}
		if err := l.svcCtx.ContactApplyModel.SoftDeleteByUserOrContact(l.ctx, now, uuid, uuid); err != nil {
			logx.Errorf("delete user %s applies error: %v", uuid, err)
			return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
		}
	}
	return &core.BaseResp{Code: 200, Message: "删除用户成功"}, nil
}
