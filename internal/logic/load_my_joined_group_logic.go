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

type LoadMyJoinedGroupLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoadMyJoinedGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoadMyJoinedGroupLogic {
	return &LoadMyJoinedGroupLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoadMyJoinedGroupLogic) LoadMyJoinedGroup(req *types.OwnlistReq) (resp *types.CommonResp, err error) {
	r, e := l.svcCtx.Core.LoadMyJoinedGroup(l.ctx, &core_client.OwnlistReq{
		OwnerId: req.OwnerId,
	})
	if e != nil {
		return &types.CommonResp{Code: 500, Message: "系统错误，请联系工作人员"}, nil
	}
	resp = &types.CommonResp{Code: int(r.Code), Message: r.Message}
	if len(r.Groups) > 0 {
		resp.Data = r.Groups
	}
	return resp, nil
}
