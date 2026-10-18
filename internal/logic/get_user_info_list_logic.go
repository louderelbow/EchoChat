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

type GetUserInfoListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserInfoListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoListLogic {
	return &GetUserInfoListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserInfoListLogic) GetUserInfoList(req *types.GetUserInfoListReq) (resp *types.CommonResp, err error) {
	r, e := l.svcCtx.Core.GetUserInfoList(l.ctx, &core_client.GetUserInfoListReq{
		OwnerId: req.OwnerId,
	})
	if e != nil {
		return &types.CommonResp{Code: 500, Message: "系统错误，请联系工作人员"}, nil
	}
	resp = &types.CommonResp{Code: int(r.Code), Message: r.Message}
	if len(r.Users) > 0 {
		resp.Data = r.Users
	}
	return resp, nil
}
