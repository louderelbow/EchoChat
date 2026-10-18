package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"echochat/pkg/constants"
	"echochat/pkg/enum"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"
)

type CancelBlackContactLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelBlackContactLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelBlackContactLogic {
	return &CancelBlackContactLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CancelBlackContactLogic) CancelBlackContact(in *core.ContactPairReq) (*core.BaseResp, error) {
	blackContact, err := l.svcCtx.UserContactModel.FindOneByUserAndContact(l.ctx, in.OwnerId, in.ContactId)
	if err != nil {
		l.Errorf(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	if blackContact.Status != int64(enum.ContactBlack) {
		return &core.BaseResp{Code: 400, Message: "未拉黑该联系人，无需解除拉黑"}, nil
	}

	beBlackContact, err := l.svcCtx.UserContactModel.FindOneByUserAndContact(l.ctx, in.ContactId, in.OwnerId)
	if err != nil {
		l.Errorf(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	if beBlackContact.Status != int64(enum.ContactBeBlack) {
		return &core.BaseResp{Code: 400, Message: "该联系人未被拉黑，无需解除拉黑"}, nil
	}

	if err := l.svcCtx.UserContactModel.UpdateStatusByUserAndContact(l.ctx, int64(enum.ContactNormal), in.OwnerId, in.ContactId); err != nil {
		l.Errorf(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	if err := l.svcCtx.UserContactModel.UpdateStatusByUserAndContact(l.ctx, int64(enum.ContactNormal), in.ContactId, in.OwnerId); err != nil {
		l.Errorf(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	return &core.BaseResp{Code: 200, Message: "已解除拉黑该联系人"}, nil
}
