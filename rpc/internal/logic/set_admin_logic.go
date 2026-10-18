package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"echochat/pkg/constants"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"
)

type SetAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetAdminLogic {
	return &SetAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetAdminLogic) SetAdmin(in *core.AbleUsersReq) (*core.BaseResp, error) {
	for _, uuid := range in.UuidList {
		if err := l.svcCtx.UserInfoModel.UpdateAdminByUuid(l.ctx, int64(in.IsAdmin), uuid); err != nil {
			l.Errorf(err.Error())
			return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
		}
	}
	return &core.BaseResp{Code: 200, Message: "设置管理员成功"}, nil
}
