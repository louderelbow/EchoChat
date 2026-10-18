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

type OpenSessionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOpenSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OpenSessionLogic {
	return &OpenSessionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OpenSessionLogic) OpenSession(req *types.OpenSessionReq) (resp *types.CommonResp, err error) {
	r, e := l.svcCtx.Core.OpenSession(l.ctx, &core_client.OpenSessionReq{
		SendId:    req.SendId,
		ReceiveId: req.ReceiveId,
	})
	if e != nil {
		return &types.CommonResp{Code: 500, Message: "系统错误，请联系工作人员"}, nil
	}
	resp = &types.CommonResp{Code: int(r.Code), Message: r.Message}
	if r.SessionId != "" {
		resp.Data = r.SessionId
	}
	return resp, nil
}
