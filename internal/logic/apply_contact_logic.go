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

type ApplyContactLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewApplyContactLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApplyContactLogic {
	return &ApplyContactLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ApplyContactLogic) ApplyContact(req *types.ApplyContactReq) (resp *types.CommonResp, err error) {
	r, e := l.svcCtx.Core.ApplyContact(l.ctx, &core_client.ApplyContactReq{
		OwnerId:   req.OwnerId,
		ContactId: req.ContactId,
		Message:   req.Message,
	})
	if e != nil {
		return &types.CommonResp{Code: 500, Message: "系统错误，请联系工作人员"}, nil
	}
	resp = &types.CommonResp{Code: int(r.Code), Message: r.Message}
	return resp, nil
}
