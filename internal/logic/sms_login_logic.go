// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"echochat/internal/svc"
	"echochat/internal/types"
	"echochat/pkg/authtoken"
	"echochat/rpc/core_client"

	"github.com/zeromicro/go-zero/core/logx"
)

type SmsLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSmsLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SmsLoginLogic {
	return &SmsLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SmsLoginLogic) SmsLogin(req *types.SmsLoginReq) (resp *types.CommonResp, err error) {
	r, e := l.svcCtx.Core.SmsLogin(l.ctx, &core_client.SmsLoginReq{
		Telephone: req.Telephone,
		SmsCode:   req.SmsCode,
	})
	if e != nil {
		return &types.CommonResp{Code: 500, Message: "系统错误，请联系工作人员"}, nil
	}
	resp = &types.CommonResp{Code: int(r.Code), Message: r.Message}
	if r.User != nil {
		resp.Data = r.User
		// 登录成功签发令牌
		if r.Code == 200 {
			if token, terr := authtoken.Issue(r.User.Uuid, r.User.IsAdmin == 1); terr == nil {
				resp.Token = token
			} else {
				l.Errorf(terr.Error())
			}
		}
	}
	return resp, nil
}
