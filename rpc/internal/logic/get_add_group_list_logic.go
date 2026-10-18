package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"echochat/pkg/constants"
	"echochat/pkg/enum"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"
)

type GetAddGroupListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAddGroupListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAddGroupListLogic {
	return &GetAddGroupListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetAddGroupListLogic) GetAddGroupList(in *core.AddGroupListReq) (*core.ContactApplyListResp, error) {
	applyList, err := l.svcCtx.ContactApplyModel.FindByContactIdAndStatus(l.ctx, in.GroupId, int64(enum.ApplyPending))
	if err != nil {
		l.Errorf(err.Error())
		return &core.ContactApplyListResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	items := make([]*core.ContactApplyItem, 0, len(applyList))
	for _, apply := range applyList {
		message := "申请理由：无"
		if apply.Message.String != "" {
			message = "申请理由：" + apply.Message.String
		}
		user, err := l.svcCtx.UserInfoModel.FindOneBriefByUuid(l.ctx, apply.UserId)
		if err != nil {
			l.Errorf(err.Error())
			return &core.ContactApplyListResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
		}
		items = append(items, &core.ContactApplyItem{
			ContactId:     user.Uuid,
			ContactName:   user.Nickname,
			ContactAvatar: user.Avatar,
			Message:       message,
		})
	}
	return &core.ContactApplyListResp{Code: 200, Message: "获取成功", Items: items}, nil
}
