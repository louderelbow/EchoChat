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

type GetGroupMessageListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetGroupMessageListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupMessageListLogic {
	return &GetGroupMessageListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetGroupMessageListLogic) GetGroupMessageList(req *types.GetGroupMessageListReq) (resp *types.CommonResp, err error) {
	r, e := l.svcCtx.Core.GetGroupMessageList(l.ctx, &core_client.GetGroupMessageListReq{
		GroupId: req.GroupId,
	})
	if e != nil {
		return &types.CommonResp{Code: 500, Message: "系统错误，请联系工作人员"}, nil
	}
	resp = &types.CommonResp{Code: int(r.Code), Message: r.Message}
	if len(r.Messages) > 0 {
		resp.Data = r.Messages
	}
	return resp, nil
}
