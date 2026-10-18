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

type GetContactInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetContactInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetContactInfoLogic {
	return &GetContactInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetContactInfoLogic) GetContactInfo(req *types.GetContactInfoReq) (resp *types.CommonResp, err error) {
	r, e := l.svcCtx.Core.GetContactInfo(l.ctx, &core_client.GetContactInfoReq{
		ContactId: req.ContactId,
	})
	if e != nil {
		return &types.CommonResp{Code: 500, Message: "系统错误，请联系工作人员"}, nil
	}
	resp = &types.CommonResp{Code: int(r.Code), Message: r.Message}
	if r.Contact != nil {
		resp.Data = r.Contact
	}
	return resp, nil
}
