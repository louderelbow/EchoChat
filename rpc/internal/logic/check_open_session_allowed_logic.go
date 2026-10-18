package logic

import (
	"context"

	"echochat/pkg/constants"
	"echochat/pkg/enum"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckOpenSessionAllowedLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckOpenSessionAllowedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckOpenSessionAllowedLogic {
	return &CheckOpenSessionAllowedLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CheckOpenSessionAllowedLogic) CheckOpenSessionAllowed(in *core.CreateSessionReq) (*core.CheckOpenSessionAllowedResp, error) {
	sendId := in.SendId
	receiveId := in.ReceiveId

	contact, err := l.svcCtx.UserContactModel.FindOneAllByUserAndContact(l.ctx, sendId, receiveId)
	if err != nil {
		l.Error(err.Error())
		return &core.CheckOpenSessionAllowedResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	if contact.Status == int64(enum.ContactBeBlack) {
		return &core.CheckOpenSessionAllowedResp{Code: 400, Message: "已被对方拉黑，无法发起会话"}, nil
	} else if contact.Status == int64(enum.ContactBlack) {
		return &core.CheckOpenSessionAllowedResp{Code: 400, Message: "已拉黑对方，先解除拉黑状态才能发起会话"}, nil
	}

	if len(receiveId) > 0 && receiveId[0] == 'U' {
		user, err := l.svcCtx.UserInfoModel.FindOneAllByUuid(l.ctx, receiveId)
		if err != nil {
			l.Error(err.Error())
			return &core.CheckOpenSessionAllowedResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
		}
		if user.Status == int64(enum.UserDisable) {
			l.Error("对方已被禁用，无法发起会话")
			return &core.CheckOpenSessionAllowedResp{Code: 400, Message: "对方已被禁用，无法发起会话"}, nil
		}
	} else {
		group, err := l.svcCtx.GroupInfoModel.FindOneAllByUuid(l.ctx, receiveId)
		if err != nil {
			l.Error(err.Error())
			return &core.CheckOpenSessionAllowedResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
		}
		if group.Status == int64(enum.GroupDisable) {
			l.Error("对方已被禁用，无法发起会话")
			return &core.CheckOpenSessionAllowedResp{Code: 400, Message: "对方已被禁用，无法发起会话"}, nil
		}
	}
	return &core.CheckOpenSessionAllowedResp{Code: 200, Message: "可以发起会话", Allowed: true}, nil
}
