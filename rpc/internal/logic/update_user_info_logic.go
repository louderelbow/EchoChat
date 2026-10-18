package logic

import (
	"context"
	"database/sql"

	"github.com/zeromicro/go-zero/core/logx"

	"echochat/pkg/constants"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"
)

type UpdateUserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserInfoLogic {
	return &UpdateUserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateUserInfoLogic) UpdateUserInfo(in *core.UpdateUserInfoReq) (*core.BaseResp, error) {
	user, err := l.svcCtx.UserInfoModel.FindOneEditableByUuid(l.ctx, in.Uuid)
	if err != nil {
		l.Errorf(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	if in.Email != "" {
		user.Email = sql.NullString{String: in.Email, Valid: true}
	}
	if in.Nickname != "" {
		user.Nickname = in.Nickname
	}
	if in.Birthday != "" {
		user.Birthday = sql.NullString{String: in.Birthday, Valid: true}
	}
	if in.Signature != "" {
		user.Signature = sql.NullString{String: in.Signature, Valid: true}
	}
	if in.Avatar != "" {
		user.Avatar = in.Avatar
	}
	// 性别始终按提交值落库：0.未知，1.男，2.女
	user.Gender = int64(in.Gender)

	err = l.svcCtx.UserInfoModel.UpdateUserInfo(l.ctx, user.Email, user.Nickname, user.Gender, user.Birthday, user.Signature, user.Avatar, user.Uuid)
	if err != nil {
		l.Errorf(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	return &core.BaseResp{Code: 200, Message: "修改用户信息成功"}, nil
}
