package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"echochat/pkg/constants"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"
)

type GetUserInfoListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserInfoListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoListLogic {
	return &GetUserInfoListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserInfoListLogic) GetUserInfoList(in *core.GetUserInfoListReq) (*core.GetUserInfoListResp, error) {
	// 管理员查询所有用户，包含已被软删除的用户（Unscoped）
	users, err := l.svcCtx.UserInfoModel.FindAllExceptUuid(l.ctx, in.OwnerId)
	if err != nil {
		l.Errorf(err.Error())
		return &core.GetUserInfoListResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	items := make([]*core.UserListItem, 0, len(users))
	for _, user := range users {
		items = append(items, &core.UserListItem{
			Uuid:      user.Uuid,
			Nickname:  user.Nickname,
			Telephone: user.Telephone,
			Status:    int32(user.Status),
			IsAdmin:   int32(user.IsAdmin),
			IsDeleted: user.DeletedAt.Valid,
		})
	}
	return &core.GetUserInfoListResp{Code: 200, Message: "获取用户列表成功", Users: items}, nil
}
