package logic

import (
	"context"

	"echochat/pkg/constants"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetGroupInfoListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetGroupInfoListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupInfoListLogic {
	return &GetGroupInfoListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetGroupInfoListLogic) GetGroupInfoList(in *core.OwnlistReq) (*core.GetGroupInfoListResp, error) {
	// 管理员列表不使用 redis 缓存，且需要包含已软删除的群聊（Unscoped）
	groupList, err := l.svcCtx.GroupInfoModel.FindAll(l.ctx)
	if err != nil {
		l.Error(err.Error())
		return &core.GetGroupInfoListResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	groups := make([]*core.GroupListItem, 0, len(groupList))
	for _, group := range groupList {
		groups = append(groups, &core.GroupListItem{
			Uuid:      group.Uuid,
			Name:      group.Name,
			OwnerId:   group.OwnerId,
			Status:    int32(group.Status),
			IsDeleted: group.DeletedAt.Valid,
		})
	}
	return &core.GetGroupInfoListResp{Code: 200, Message: "获取成功", Groups: groups}, nil
}
