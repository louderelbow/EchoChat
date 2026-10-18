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

type CheckOpenSessionAllowedLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCheckOpenSessionAllowedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckOpenSessionAllowedLogic {
	return &CheckOpenSessionAllowedLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CheckOpenSessionAllowedLogic) CheckOpenSessionAllowed(req *types.CreateSessionReq) (resp *types.CommonResp, err error) {
	r, e := l.svcCtx.Core.CheckOpenSessionAllowed(l.ctx, &core_client.CreateSessionReq{
		SendId:    req.SendId,
		ReceiveId: req.ReceiveId,
	})
	if e != nil {
		return &types.CommonResp{Code: 500, Message: "系统错误，请联系工作人员"}, nil
	}
	resp = &types.CommonResp{Code: int(r.Code), Message: r.Message, Data: r.Allowed}
	return resp, nil
}
