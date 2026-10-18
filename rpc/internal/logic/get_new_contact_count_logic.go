package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"echochat/pkg/constants"
	"echochat/pkg/enum"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"
)

type GetNewContactCountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetNewContactCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNewContactCountLogic {
	return &GetNewContactCountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetNewContactCount 统计当前用户待处理的申请总数，供前端红点提示：
//  1. 好友申请：contact_id = 当前用户（别人申请加我为好友）
//  2. 加群申请：contact_id 属于「我是群主」的群聊
func (l *GetNewContactCountLogic) GetNewContactCount(in *core.OwnlistReq) (*core.ContactApplyCountResp, error) {
	ownerId := in.OwnerId

	friendCount, err := l.svcCtx.ContactApplyModel.CountByContactIdAndStatus(l.ctx, ownerId, int64(enum.ApplyPending))
	if err != nil {
		l.Errorf(err.Error())
		return &core.ContactApplyCountResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	groupCount := int64(0)
	groups, gerr := l.svcCtx.GroupInfoModel.FindByOwnerId(l.ctx, ownerId)
	if gerr != nil {
		l.Errorf(gerr.Error())
		return &core.ContactApplyCountResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	if len(groups) > 0 {
		groupIds := make([]string, 0, len(groups))
		for _, g := range groups {
			groupIds = append(groupIds, g.Uuid)
		}
		groupCount, err = l.svcCtx.ContactApplyModel.CountByContactIdsAndStatus(l.ctx, groupIds, int64(enum.ApplyPending))
		if err != nil {
			l.Errorf(err.Error())
			return &core.ContactApplyCountResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
		}
	}

	return &core.ContactApplyCountResp{Code: 200, Message: "获取成功", Count: friendCount + groupCount}, nil
}
