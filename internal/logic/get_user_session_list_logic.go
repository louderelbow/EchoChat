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

type GetUserSessionListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserSessionListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserSessionListLogic {
	return &GetUserSessionListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserSessionListLogic) GetUserSessionList(req *types.OwnlistReq) (resp *types.CommonResp, err error) {
	r, e := l.svcCtx.Core.GetUserSessionList(l.ctx, &core_client.OwnlistReq{
		OwnerId: req.OwnerId,
	})
	if e != nil {
		return &types.CommonResp{Code: 500, Message: "系统错误，请联系工作人员"}, nil
	}
	resp = &types.CommonResp{Code: int(r.Code), Message: r.Message}
	if len(r.Sessions) > 0 {
		resp.Data = r.Sessions
	}
	return resp, nil
}
