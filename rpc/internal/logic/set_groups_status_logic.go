package logic

import (
	"context"

	"echochat/pkg/constants"
	"echochat/pkg/enum"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetGroupsStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetGroupsStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetGroupsStatusLogic {
	return &SetGroupsStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetGroupsStatusLogic) SetGroupsStatus(in *core.SetGroupsStatusReq) (*core.BaseResp, error) {
	status := in.Status
	for _, uuid := range in.UuidList {
		if err := l.svcCtx.GroupInfoModel.UpdateStatusByUuid(l.ctx, int64(status), uuid); err != nil {
			l.Error(err.Error())
			return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
		}
		if status == int32(enum.GroupDisable) {
			if err := l.svcCtx.SessionModel.DismissNowByReceiveId(l.ctx, uuid); err != nil {
				l.Error(err.Error())
				return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
			}
		}
	}
	return &core.BaseResp{Code: 200, Message: "设置成功"}, nil
}
