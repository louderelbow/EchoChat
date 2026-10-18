package logic

import (
	"context"
	"errors"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"echochat/pkg/constants"
	"echochat/pkg/redisx"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"
)

type SmsLoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSmsLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SmsLoginLogic {
	return &SmsLoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SmsLoginLogic) SmsLogin(in *core.SmsLoginReq) (*core.LoginResp, error) {
	user, err := l.svcCtx.UserInfoModel.FindOneByTelephoneNoPassword(l.ctx, in.Telephone)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return &core.LoginResp{Code: 400, Message: "用户不存在，请注册"}, nil
		}
		l.Errorf(err.Error())
		return &core.LoginResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	key := "auth_code_" + in.Telephone
	code, err := redisx.GetKey(key)
	if err != nil {
		l.Errorf(err.Error())
		return &core.LoginResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	if code != in.SmsCode {
		return &core.LoginResp{Code: 400, Message: "验证码不正确，请重试"}, nil
	}
	if err := redisx.DelKeyIfExists(key); err != nil {
		l.Errorf(err.Error())
		return &core.LoginResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	year, month, day := user.CreatedAt.Date()
	return &core.LoginResp{
		Code:    200,
		Message: "登陆成功",
		User: &core.UserInfo{
			Uuid:      user.Uuid,
			Nickname:  user.Nickname,
			Telephone: user.Telephone,
			Email:     user.Email.String,
			Avatar:    user.Avatar,
			Gender:    int32(user.Gender),
			Birthday:  user.Birthday.String,
			Signature: user.Signature.String,
			CreatedAt: fmt.Sprintf("%d.%d.%d", year, month, day),
			IsAdmin:   int32(user.IsAdmin),
			Status:    int32(user.Status),
		},
	}, nil
}
