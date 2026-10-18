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

type DeleteGroupsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteGroupsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteGroupsLogic {
	return &DeleteGroupsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteGroupsLogic) DeleteGroups(req *types.DeleteGroupsReq) (resp *types.CommonResp, err error) {
	r, e := l.svcCtx.Core.DeleteGroups(l.ctx, &core_client.DeleteGroupsReq{
		UuidList: req.UuidList,
	})
	if e != nil {
		return &types.CommonResp{Code: 500, Message: "系统错误，请联系工作人员"}, nil
	}
	resp = &types.CommonResp{Code: int(r.Code), Message: r.Message}
	return resp, nil
}
