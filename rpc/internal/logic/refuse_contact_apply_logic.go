package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"echochat/pkg/constants"
	"echochat/pkg/enum"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"
)

type RefuseContactApplyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRefuseContactApplyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefuseContactApplyLogic {
	return &RefuseContactApplyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RefuseContactApplyLogic) RefuseContactApply(in *core.ContactPairReq) (*core.BaseResp, error) {
	apply, err := l.svcCtx.ContactApplyModel.FindOneBasicByContactUser(l.ctx, in.OwnerId, in.ContactId)
	if err != nil {
		l.Errorf(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	if err := l.svcCtx.ContactApplyModel.UpdateStatusByUuid(l.ctx, int64(enum.ApplyRefuse), apply.Uuid); err != nil {
		l.Errorf(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	if len(in.OwnerId) > 0 && in.OwnerId[0] == 'U' {
		return &core.BaseResp{Code: 200, Message: "已拒绝该联系人申请"}, nil
	}
	return &core.BaseResp{Code: 200, Message: "已拒绝该加群申请"}, nil
}
