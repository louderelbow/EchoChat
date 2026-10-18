package logic

import (
	"context"
	"encoding/json"
	"errors"

	"echochat/pkg/constants"
	"echochat/pkg/redisx"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetGroupMemberListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetGroupMemberListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupMemberListLogic {
	return &GetGroupMemberListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetGroupMemberListLogic) GetGroupMemberList(in *core.GetGroupMemberListReq) (*core.GetGroupMemberListResp, error) {
	groupId := in.GroupId
	rspString, err := redisx.GetKeyNilIsErr("group_memberlist_" + groupId)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			group, qerr := l.svcCtx.GroupInfoModel.FindOneAllByUuid(l.ctx, groupId)
			if qerr != nil {
				l.Error(qerr.Error())
				return &core.GetGroupMemberListResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
			}
			var members []string
			if uerr := json.Unmarshal([]byte(group.Members.String), &members); uerr != nil {
				l.Error(uerr.Error())
				return &core.GetGroupMemberListResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
			}
			memberList := make([]*core.GroupMember, 0, len(members))
			for _, member := range members {
				user, qerr := l.svcCtx.UserInfoModel.FindOneAllByUuid(l.ctx, member)
				if qerr != nil {
					l.Error(qerr.Error())
					return &core.GetGroupMemberListResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
				}
				memberList = append(memberList, &core.GroupMember{
					UserId:   user.Uuid,
					Nickname: user.Nickname,
					Avatar:   user.Avatar,
				})
			}
			return &core.GetGroupMemberListResp{Code: 200, Message: "获取群聊成员列表成功", Members: memberList}, nil
		}
		l.Error(err.Error())
		return &core.GetGroupMemberListResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	var cached []*core.GroupMember
	if err := json.Unmarshal([]byte(rspString), &cached); err != nil {
		l.Error(err.Error())
	}
	return &core.GetGroupMemberListResp{Code: 200, Message: "获取群聊成员列表成功", Members: cached}, nil
}
