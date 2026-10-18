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

type GetNewContactCountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetNewContactCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNewContactCountLogic {
	return &GetNewContactCountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetNewContactCount 返回当前用户待处理申请数量（好友申请 + 我作为群主的加群申请）。
func (l *GetNewContactCountLogic) GetNewContactCount(req *types.OwnlistReq) (resp *types.CommonResp, err error) {
	r, e := l.svcCtx.Core.GetNewContactCount(l.ctx, &core_client.OwnlistReq{
		OwnerId: req.OwnerId,
	})
	if e != nil {
		return &types.CommonResp{Code: 500, Message: "系统错误，请联系工作人员"}, nil
	}
	return &types.CommonResp{Code: int(r.Code), Message: r.Message, Data: r.Count}, nil
}
