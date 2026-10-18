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

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterReq) (resp *types.CommonResp, err error) {
	r, e := l.svcCtx.Core.Register(l.ctx, &core_client.RegisterReq{
		Telephone: req.Telephone,
		Password:  req.Password,
		Nickname:  req.Nickname,
		SmsCode:   req.SmsCode,
	})
	if e != nil {
		return &types.CommonResp{Code: 500, Message: "系统错误，请联系工作人员"}, nil
	}
	resp = &types.CommonResp{Code: int(r.Code), Message: r.Message}
	if r.User != nil {
		resp.Data = r.User
		// 注册成功签发令牌
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
