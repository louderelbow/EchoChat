package logic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"echochat/internal/model"
	"echochat/pkg/constants"
	"echochat/pkg/enum"
	"echochat/pkg/password"
	"echochat/pkg/random"
	"echochat/pkg/redisx"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegisterLogic) Register(in *core.RegisterReq) (*core.RegisterResp, error) {
	key := "auth_code_" + in.Telephone
	code, err := redisx.GetKey(key)
	if err != nil {
		l.Errorf(err.Error())
		return &core.RegisterResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	if code != in.SmsCode {
		return &core.RegisterResp{Code: 400, Message: "验证码不正确，请重试"}, nil
	}
	if err := redisx.DelKeyIfExists(key); err != nil {
		l.Errorf(err.Error())
		return &core.RegisterResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	// 判断邮箱是否已经被注册过了
	_, err = l.svcCtx.UserInfoModel.FindIdByTelephone(l.ctx, in.Telephone)
	if err == nil {
		return &core.RegisterResp{Code: 400, Message: "该邮箱已经存在，注册失败"}, nil
	}
	if !errors.Is(err, sqlx.ErrNotFound) {
		l.Errorf(err.Error())
		return &core.RegisterResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	// 密码以 bcrypt 哈希存储，禁止明文落库
	hashedPassword, herr := password.Hash(in.Password)
	if herr != nil {
		l.Errorf(herr.Error())
		return &core.RegisterResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	newUser := &model.UserInfo{
		Uuid:      "U" + random.GetNowAndLenRandomString(11),
		Telephone: in.Telephone,
		Password:  hashedPassword,
		Nickname:  in.Nickname,
		Avatar:    "https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png",
		CreatedAt: time.Now(),
		IsAdmin:   0,
		Status:    int64(enum.UserNormal),
	}
	err = l.svcCtx.UserInfoModel.InsertRegisterUser(l.ctx, newUser.Uuid, newUser.Nickname, newUser.Telephone, "", newUser.Avatar, 0, "", newUser.Password, "", 0, int64(enum.UserNormal), newUser.CreatedAt)
	if err != nil {
		l.Errorf(err.Error())
		return &core.RegisterResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	year, month, day := newUser.CreatedAt.Date()
	return &core.RegisterResp{
		Code:    200,
		Message: "注册成功",
		User: &core.UserInfo{
			Uuid:      newUser.Uuid,
			Nickname:  newUser.Nickname,
			Telephone: newUser.Telephone,
			Email:     "",
			Avatar:    newUser.Avatar,
			Gender:    0,
			Birthday:  "",
			Signature: "",
			CreatedAt: fmt.Sprintf("%d.%d.%d", year, month, day),
			IsAdmin:   0,
			Status:    int32(enum.UserNormal),
		},
	}, nil
}
