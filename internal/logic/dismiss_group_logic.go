// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"echochat/internal/middleware"
	"echochat/internal/svc"
	"echochat/internal/types"
	"echochat/rpc/core_client"

	"github.com/zeromicro/go-zero/core/logx"
)

type DismissGroupLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDismissGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DismissGroupLogic {
	return &DismissGroupLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DismissGroupLogic) DismissGroup(req *types.DismissGroupReq) (resp *types.CommonResp, err error) {
	// 群解散只允许群主本人（开启鉴权时校验调用者身份）
	if !middleware.IsOwner(l.ctx, req.OwnerId) {
		return &types.CommonResp{Code: 403, Message: "无权限操作"}, nil
	}
	r, e := l.svcCtx.Core.DismissGroup(l.ctx, &core_client.DismissGroupReq{
		OwnerId: req.OwnerId,
		GroupId: req.GroupId,
	})
	if e != nil {
		return &types.CommonResp{Code: 500, Message: "系统错误，请联系工作人员"}, nil
	}
	resp = &types.CommonResp{Code: int(r.Code), Message: r.Message}
	return resp, nil
}
