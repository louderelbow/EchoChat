package logic

import (
	"context"
	"encoding/json"

	"echochat/pkg/constants"
	"echochat/pkg/enum"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type LeaveGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLeaveGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LeaveGroupLogic {
	return &LeaveGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LeaveGroupLogic) LeaveGroup(in *core.LeaveGroupReq) (*core.BaseResp, error) {
	userId := in.UserId
	groupId := in.GroupId

	// 从群聊中清除该用户
	group, err := l.svcCtx.GroupInfoModel.FindOneAllByUuid(l.ctx, groupId)
	if err != nil {
		l.Error(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	var members []string
	if err := json.Unmarshal([]byte(group.Members.String), &members); err != nil {
		l.Error(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	for i, member := range members {
		if member == userId {
			members = append(members[:i], members[i+1:]...)
			break
		}
	}
	data, err := json.Marshal(members)
	if err != nil {
		l.Error(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	if err := l.svcCtx.GroupInfoModel.UpdateMembersAndCntNoTimeByUuid(l.ctx, string(data), group.MemberCnt-1, groupId); err != nil {
		l.Error(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	// 删除会话
	if err := l.svcCtx.SessionModel.DismissNowBySendReceive(l.ctx, userId, groupId); err != nil {
		l.Error(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	// 删除联系人，并标记为退群
	if err := l.svcCtx.UserContactModel.SoftDeleteNowAndStatusByUserAndContact(l.ctx, int64(enum.ContactQuitGroup), userId, groupId); err != nil {
		l.Error(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	// 删除申请记录
	if err := l.svcCtx.ContactApplyModel.SoftDeleteNowByContactUser(l.ctx, groupId, userId); err != nil {
		l.Error(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	return &core.BaseResp{Code: 200, Message: "退群成功"}, nil
}
