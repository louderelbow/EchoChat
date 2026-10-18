package logic

import (
	"context"
	"encoding/json"
	"errors"

	redis "github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"echochat/pkg/constants"
	"echochat/pkg/redisx"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"
)

type GetUserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserInfoLogic) GetUserInfo(in *core.GetUserInfoReq) (*core.GetUserInfoResp, error) {
	rspString, err := redisx.GetKeyNilIsErr("user_info_" + in.Uuid)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			user, e := l.svcCtx.UserInfoModel.FindOneProfileByUuid(l.ctx, in.Uuid)
			if e != nil && !errors.Is(e, sqlx.ErrNotFound) {
				l.Errorf(e.Error())
				return &core.GetUserInfoResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
			}
			return &core.GetUserInfoResp{
				Code:    200,
				Message: "获取用户信息成功",
				User: &core.UserInfo{
					Uuid:      user.Uuid,
					Nickname:  user.Nickname,
					Telephone: user.Telephone,
					Email:     user.Email.String,
					Avatar:    user.Avatar,
					Gender:    int32(user.Gender),
					Birthday:  user.Birthday.String,
					Signature: user.Signature.String,
					CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
					IsAdmin:   int32(user.IsAdmin),
					Status:    int32(user.Status),
				},
			}, nil
		}
		l.Errorf(err.Error())
		return &core.GetUserInfoResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	var rsp core.UserInfo
	if err := json.Unmarshal([]byte(rspString), &rsp); err != nil {
		l.Errorf(err.Error())
	}
	return &core.GetUserInfoResp{Code: 200, Message: "获取用户信息成功", User: &rsp}, nil
}
