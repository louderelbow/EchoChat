package logic

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	redis "github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"

	"echochat/pkg/constants"
	"echochat/pkg/enum"
	"echochat/pkg/redisx"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"
)

type LoadMyJoinedGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoadMyJoinedGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoadMyJoinedGroupLogic {
	return &LoadMyJoinedGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoadMyJoinedGroupLogic) LoadMyJoinedGroup(in *core.OwnlistReq) (*core.LoadMyJoinedGroupResp, error) {
	rspString, err := redisx.GetKeyNilIsErr("my_joined_group_list_" + in.OwnerId)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			contactList, e := l.svcCtx.UserContactModel.FindByUserIdExcludeTwoStatus(l.ctx, in.OwnerId, int64(enum.ContactQuitGroup), int64(enum.ContactKickOutGroup))
			if e != nil {
				l.Errorf(e.Error())
				return &core.LoadMyJoinedGroupResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
			}

			var groupListRsp []*core.GroupBriefItem
			for _, contact := range contactList {
				if len(contact.ContactId) > 0 && contact.ContactId[0] == 'G' {
					group, e := l.svcCtx.GroupInfoModel.FindOneBriefByUuid(l.ctx, contact.ContactId)
					if e != nil {
						l.Errorf(e.Error())
						return &core.LoadMyJoinedGroupResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
					}
					// 群主不是自己
					if group.OwnerId != in.OwnerId {
						groupListRsp = append(groupListRsp, &core.GroupBriefItem{
							GroupId:   group.Uuid,
							GroupName: group.Name,
							Avatar:    defaultAvatarIfEmpty(group.Avatar),
						})
					}
				}
			}

			if b, e := json.Marshal(groupListRsp); e != nil {
				l.Errorf(e.Error())
			} else if e := redisx.SetKeyEx("my_joined_group_list_"+in.OwnerId, string(b), time.Minute*time.Duration(constants.REDIS_TIMEOUT)); e != nil {
				l.Errorf(e.Error())
			}
			return &core.LoadMyJoinedGroupResp{Code: 200, Message: "获取加入群成功", Groups: groupListRsp}, nil
		}
		l.Errorf(err.Error())
		return &core.LoadMyJoinedGroupResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	var rsp []*core.GroupBriefItem
	if err := json.Unmarshal([]byte(rspString), &rsp); err != nil {
		l.Errorf(err.Error())
	}
	return &core.LoadMyJoinedGroupResp{Code: 200, Message: "获取加入群成功", Groups: rsp}, nil
}
