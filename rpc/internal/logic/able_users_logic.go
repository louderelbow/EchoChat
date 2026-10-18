package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"echochat/pkg/constants"
	"echochat/pkg/enum"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"
)

type AbleUsersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAbleUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AbleUsersLogic {
	return &AbleUsersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AbleUsersLogic) AbleUsers(in *core.AbleUsersReq) (*core.BaseResp, error) {
	for _, uuid := range in.UuidList {
		if err := l.svcCtx.UserInfoModel.UpdateStatusByUuid(l.ctx, int64(enum.UserNormal), uuid); err != nil {
			l.Errorf(err.Error())
			return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
		}
	}
	// 用户是否启用禁用需要实时更新 contact_user_list 状态，原实现中删除 redis 缓存的代码被注释
	return &core.BaseResp{Code: 200, Message: "启用用户成功"}, nil
}
