package logic

import (
	"context"
	"encoding/json"

	"echochat/pkg/constants"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemoveGroupMembersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRemoveGroupMembersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveGroupMembersLogic {
	return &RemoveGroupMembersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RemoveGroupMembersLogic) RemoveGroupMembers(in *core.RemoveGroupMembersReq) (*core.BaseResp, error) {
	group, err := l.svcCtx.GroupInfoModel.FindOneAllByUuid(l.ctx, in.GroupId)
	if err != nil {
		l.Error(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	var members []string
	if err := json.Unmarshal([]byte(group.Members.String), &members); err != nil {
		l.Error(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	memberCnt := group.MemberCnt
	for _, uuid := range in.UuidList {
		if in.OwnerId == uuid {
			return &core.BaseResp{Code: 400, Message: "不能移除群主"}, nil
		}
		// 从 members 中找到 uuid，移除
		for i, member := range members {
			if member == uuid {
				members = append(members[:i], members[i+1:]...)
			}
		}
		memberCnt -= 1

		// 删除会话
		if err := l.svcCtx.SessionModel.DismissNowBySendReceive(l.ctx, uuid, in.GroupId); err != nil {
			l.Error(err.Error())
			return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
		}
		// 删除联系人
		if err := l.svcCtx.UserContactModel.SoftDeleteNowByUserAndContact(l.ctx, uuid, in.GroupId); err != nil {
			l.Error(err.Error())
			return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
		}
		// 删除申请记录
		if err := l.svcCtx.ContactApplyModel.SoftDeleteNowByUserContact(l.ctx, uuid, in.GroupId); err != nil {
			l.Error(err.Error())
			return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
		}
	}

	data, err := json.Marshal(members)
	if err != nil {
		l.Error(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	if err := l.svcCtx.GroupInfoModel.UpdateMembersAndCntNoTimeByUuid(l.ctx, string(data), memberCnt, in.GroupId); err != nil {
		l.Error(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	return &core.BaseResp{Code: 200, Message: "移除群聊成员成功"}, nil
}
