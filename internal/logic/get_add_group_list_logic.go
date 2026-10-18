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

type GetAddGroupListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAddGroupListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAddGroupListLogic {
	return &GetAddGroupListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAddGroupListLogic) GetAddGroupList(req *types.AddGroupListReq) (resp *types.CommonResp, err error) {
	r, e := l.svcCtx.Core.GetAddGroupList(l.ctx, &core_client.AddGroupListReq{
		GroupId: req.GroupId,
	})
	if e != nil {
		return &types.CommonResp{Code: 500, Message: "系统错误，请联系工作人员"}, nil
	}
	resp = &types.CommonResp{Code: int(r.Code), Message: r.Message}
	if len(r.Items) > 0 {
		resp.Data = r.Items
	}
	return resp, nil
}
