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

type UpdateGroupInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateGroupInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateGroupInfoLogic {
	return &UpdateGroupInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateGroupInfoLogic) UpdateGroupInfo(req *types.UpdateGroupInfoReq) (resp *types.CommonResp, err error) {
	// 修改群资料只允许群主（开启鉴权时校验调用者身份）
	if !middleware.IsOwner(l.ctx, req.OwnerId) {
		return &types.CommonResp{Code: 403, Message: "无权限操作"}, nil
	}
	r, e := l.svcCtx.Core.UpdateGroupInfo(l.ctx, &core_client.UpdateGroupInfoReq{
		OwnerId: req.OwnerId,
		Uuid:    req.Uuid,
		Name:    req.Name,
		Avatar:  req.Avatar,
		AddMode: int32(req.AddMode),
		Notice:  req.Notice,
	})
	if e != nil {
		return &types.CommonResp{Code: 500, Message: "系统错误，请联系工作人员"}, nil
	}
	resp = &types.CommonResp{Code: int(r.Code), Message: r.Message}
	return resp, nil
}
