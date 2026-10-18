package logic

import (
	"context"

	"echochat/pkg/constants"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DismissGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDismissGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DismissGroupLogic {
	return &DismissGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DismissGroupLogic) DismissGroup(in *core.DismissGroupReq) (*core.BaseResp, error) {
	groupId := in.GroupId

	if err := l.svcCtx.GroupInfoModel.DismissByUuid(l.ctx, groupId); err != nil {
		l.Error(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	// 删除会话
	if err := l.svcCtx.SessionModel.DismissNowByReceiveId(l.ctx, groupId); err != nil {
		l.Error(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	// 删除联系人
	if err := l.svcCtx.UserContactModel.SoftDeleteNowByContactId(l.ctx, groupId); err != nil {
		l.Error(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	// 删除申请记录
	if err := l.svcCtx.ContactApplyModel.SoftDeleteNowByContactId(l.ctx, groupId); err != nil {
		l.Error(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	return &core.BaseResp{Code: 200, Message: "解散群聊成功"}, nil
}
