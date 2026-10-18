// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"echochat/internal/svc"
	"echochat/internal/types"
	"echochat/rpc/core_client"

	"github.com/zeromicro/go-zero/core/logx"
)

type BlackApplyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBlackApplyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BlackApplyLogic {
	return &BlackApplyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BlackApplyLogic) BlackApply(req *types.BlackApplyReq) (resp *types.CommonResp, err error) {
	r, e := l.svcCtx.Core.BlackApply(l.ctx, &core_client.ContactPairReq{
		OwnerId:   req.OwnerId,
		ContactId: req.ContactId,
	})
	if e != nil {
		return &types.CommonResp{Code: 500, Message: "系统错误，请联系工作人员"}, nil
	}
	resp = &types.CommonResp{Code: int(r.Code), Message: r.Message}
	return resp, nil
}
