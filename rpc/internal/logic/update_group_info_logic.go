package logic

import (
	"context"
	"database/sql"

	"echochat/pkg/constants"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateGroupInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateGroupInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateGroupInfoLogic {
	return &UpdateGroupInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateGroupInfoLogic) UpdateGroupInfo(in *core.UpdateGroupInfoReq) (*core.BaseResp, error) {
	group, err := l.svcCtx.GroupInfoModel.FindOneAllByUuid(l.ctx, in.Uuid)
	if err != nil {
		l.Error(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	name := group.Name
	if in.Name != "" {
		name = in.Name
	}
	addMode := group.AddMode
	if in.AddMode != -1 {
		addMode = int64(in.AddMode)
	}
	notice := group.Notice
	if in.Notice != "" {
		notice = sql.NullString{String: in.Notice, Valid: true}
	}
	avatar := group.Avatar
	if in.Avatar != "" {
		avatar = in.Avatar
	}

	if err := l.svcCtx.GroupInfoModel.UpdateNameAddModeNoticeAvatarByUuid(l.ctx, name, addMode, notice, avatar, in.Uuid); err != nil {
		l.Error(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	// 修改会话
	if err := l.svcCtx.SessionModel.UpdateReceiveNameAvatarByReceiveId(l.ctx, name, avatar, in.Uuid); err != nil {
		l.Error(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	return &core.BaseResp{Code: 200, Message: "更新成功"}, nil
}
