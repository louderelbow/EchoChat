package logic

import (
	"context"
	"errors"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"echochat/pkg/constants"
	"echochat/pkg/password"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 用户
func (l *LoginLogic) Login(in *core.LoginReq) (*core.LoginResp, error) {
	user, err := l.svcCtx.UserInfoModel.FindOneByTelephone(l.ctx, in.Telephone)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return &core.LoginResp{Code: 400, Message: "用户不存在，请注册"}, nil
		}
		l.Errorf(err.Error())
		return &core.LoginResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	if password.IsHashed(user.Password) {
		// 已是 bcrypt 哈希：直接校验
		if !password.Verify(user.Password, in.Password) {
			return &core.LoginResp{Code: 400, Message: "密码不正确，请重试"}, nil
		}
	} else {
		// 兼容历史明文密码：校验通过后平滑升级为 bcrypt 哈希
		if user.Password != in.Password {
			return &core.LoginResp{Code: 400, Message: "密码不正确，请重试"}, nil
		}
		if hashed, herr := password.Hash(in.Password); herr == nil {
			if uerr := l.svcCtx.UserInfoModel.UpdatePasswordByUuid(l.ctx, hashed, user.Uuid); uerr != nil {
				l.Errorf(uerr.Error())
			}
		}
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
