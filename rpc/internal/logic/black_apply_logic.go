package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"echochat/pkg/constants"
	"echochat/pkg/enum"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"
)

type BlackApplyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBlackApplyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BlackApplyLogic {
	return &BlackApplyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *BlackApplyLogic) BlackApply(in *core.ContactPairReq) (*core.BaseResp, error) {
	apply, err := l.svcCtx.ContactApplyModel.FindOneBasicByContactUser(l.ctx, in.OwnerId, in.ContactId)
	if err != nil {
		l.Errorf(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	if err := l.svcCtx.ContactApplyModel.UpdateStatusByUuid(l.ctx, int64(enum.ApplyBlack), apply.Uuid); err != nil {
		l.Errorf(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	return &core.BaseResp{Code: 200, Message: "已拉黑该申请"}, nil
}
