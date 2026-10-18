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

type DeleteContactLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteContactLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteContactLogic {
	return &DeleteContactLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteContactLogic) DeleteContact(in *core.ContactPairReq) (*core.BaseResp, error) {
	now := time.Now()
	// 自己删除
	if err := l.svcCtx.UserContactModel.UpdateDeletedAtStatusByUserAndContact(l.ctx, now, int64(enum.ContactDelete), in.OwnerId, in.ContactId); err != nil {
		l.Errorf(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	// 被删除
	if err := l.svcCtx.UserContactModel.UpdateDeletedAtStatusByUserAndContact(l.ctx, now, int64(enum.ContactBeDelete), in.ContactId, in.OwnerId); err != nil {
		l.Errorf(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	// 删除双方会话
	if err := l.svcCtx.SessionModel.UpdateDeletedAtBySendReceive(l.ctx, now, in.OwnerId, in.ContactId); err != nil {
		l.Errorf(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	if err := l.svcCtx.SessionModel.UpdateDeletedAtBySendReceive(l.ctx, now, in.ContactId, in.OwnerId); err != nil {
		l.Errorf(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	// 删除申请记录
	if err := l.svcCtx.ContactApplyModel.SoftDeleteByContactUser(l.ctx, now, in.OwnerId, in.ContactId); err != nil {
		l.Errorf(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	if err := l.svcCtx.ContactApplyModel.SoftDeleteByContactUser(l.ctx, now, in.ContactId, in.OwnerId); err != nil {
		l.Errorf(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	return &core.BaseResp{Code: 200, Message: "删除联系人成功"}, nil
}
