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

type CheckGroupAddModeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCheckGroupAddModeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckGroupAddModeLogic {
	return &CheckGroupAddModeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CheckGroupAddModeLogic) CheckGroupAddMode(req *types.CheckGroupAddModeReq) (resp *types.CommonResp, err error) {
	r, e := l.svcCtx.Core.CheckGroupAddMode(l.ctx, &core_client.CheckGroupAddModeReq{
		GroupId: req.GroupId,
	})
	if e != nil {
		return &types.CommonResp{Code: 500, Message: "系统错误，请联系工作人员"}, nil
	}
	resp = &types.CommonResp{Code: int(r.Code), Message: r.Message, Data: r.AddMode}
	return resp, nil
}
