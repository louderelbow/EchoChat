// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"echochat/internal/svc"
	"echochat/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type WsLogoutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWsLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WsLogoutLogic {
	return &WsLogoutLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WsLogoutLogic) WsLogout(req *types.WsLogoutReq) (resp *types.CommonResp, err error) {
	l.svcCtx.Hub.Remove(req.OwnerId)
	return &types.CommonResp{Code: 200, Message: "退出成功"}, nil
}
