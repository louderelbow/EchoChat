package logic

import (
	"context"

	"echochat/pkg/constants"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteGroupsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteGroupsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteGroupsLogic {
	return &DeleteGroupsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteGroupsLogic) DeleteGroups(in *core.DeleteGroupsReq) (*core.BaseResp, error) {
	for _, uuid := range in.UuidList {
		if err := l.svcCtx.GroupInfoModel.SoftDeleteByUuid(l.ctx, uuid); err != nil {
			l.Error(err.Error())
			return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
		}
		// 删除会话
		if err := l.svcCtx.SessionModel.DismissNowByReceiveId(l.ctx, uuid); err != nil {
			l.Error(err.Error())
			return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
		}
		// 删除联系人
		if err := l.svcCtx.UserContactModel.SoftDeleteNowByContactId(l.ctx, uuid); err != nil {
			l.Error(err.Error())
			return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
		}
		// 删除申请记录
		if err := l.svcCtx.ContactApplyModel.SoftDeleteNowByContactId(l.ctx, uuid); err != nil {
			l.Error(err.Error())
			return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
		}
	}

	return &core.BaseResp{Code: 200, Message: "解散/删除群聊成功"}, nil
}
