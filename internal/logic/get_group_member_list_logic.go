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

type GetGroupMemberListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetGroupMemberListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupMemberListLogic {
	return &GetGroupMemberListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetGroupMemberListLogic) GetGroupMemberList(req *types.GetGroupMemberListReq) (resp *types.CommonResp, err error) {
	r, e := l.svcCtx.Core.GetGroupMemberList(l.ctx, &core_client.GetGroupMemberListReq{
		GroupId: req.GroupId,
	})
	if e != nil {
		return &types.CommonResp{Code: 500, Message: "系统错误，请联系工作人员"}, nil
	}
	resp = &types.CommonResp{Code: int(r.Code), Message: r.Message}
	if len(r.Members) > 0 {
		resp.Data = r.Members
	}
	return resp, nil
}
