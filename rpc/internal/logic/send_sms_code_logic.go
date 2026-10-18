package logic

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"echochat/pkg/constants"
	"echochat/pkg/mail"
	"echochat/pkg/redisx"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"
)

// smsCooldown 同一邮箱发送验证码的最小间隔
const smsCooldown = 60 * time.Second

type SendSmsCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendSmsCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendSmsCodeLogic {
	return &SendSmsCodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SendSmsCodeLogic) SendSmsCode(in *core.SendSmsCodeReq) (*core.BaseResp, error) {
	// 频控：同一邮箱在冷却时间内只能发送一次，避免被刷
	cooldownKey := "sms_code_cooldown_" + in.Telephone
	if v, err := redisx.GetKey(cooldownKey); err == nil && v != "" {
		return &core.BaseResp{Code: 400, Message: "验证码发送过于频繁，请稍后再试"}, nil
	}

	message, code := mail.VerificationCode(in.Telephone)
	switch code {
	case 0:
		// 发送成功才进入冷却
		if err := redisx.SetKeyEx(cooldownKey, "1", smsCooldown); err != nil {
			l.Errorf(err.Error())
		}
		return &core.BaseResp{Code: 200, Message: message}, nil
	case -2:
		return &core.BaseResp{Code: 400, Message: message}, nil
	default:
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
}
