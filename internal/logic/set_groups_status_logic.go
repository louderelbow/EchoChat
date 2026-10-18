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

type SetGroupsStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetGroupsStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetGroupsStatusLogic {
	return &SetGroupsStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetGroupsStatusLogic) SetGroupsStatus(req *types.SetGroupsStatusReq) (resp *types.CommonResp, err error) {
	r, e := l.svcCtx.Core.SetGroupsStatus(l.ctx, &core_client.SetGroupsStatusReq{
		UuidList: req.UuidList,
		Status:   int32(req.Status),
	})
	if e != nil {
		return &types.CommonResp{Code: 500, Message: "系统错误，请联系工作人员"}, nil
	}
	resp = &types.CommonResp{Code: int(r.Code), Message: r.Message}
	return resp, nil
}
