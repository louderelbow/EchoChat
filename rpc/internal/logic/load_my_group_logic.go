package logic

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"echochat/pkg/constants"
	"echochat/pkg/redisx"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

type LoadMyGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoadMyGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoadMyGroupLogic {
	return &LoadMyGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// defaultAvatarIfEmpty 头像为空时回退到默认头像。
// 前端对头像字段做了 .startsWith("http") 处理且未判空：返回空值时该字段会因 omitempty
// 从 JSON 中消失，前端直接抛异常，进而导致整个响应/列表渲染失败，所以必须兜底。
func defaultAvatarIfEmpty(avatar string) string {
	if avatar == "" {
		return constants.DEFAULT_AVATAR
	}
	return avatar
}

func (l *LoadMyGroupLogic) LoadMyGroup(in *core.OwnlistReq) (*core.LoadMyJoinedGroupResp, error) {
	ownerId := in.OwnerId
	cacheKey := "contact_mygroup_list_" + ownerId
	rspString, err := redisx.GetKeyNilIsErr(cacheKey)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			groupList, qerr := l.svcCtx.GroupInfoModel.FindByOwnerId(l.ctx, ownerId)
			if qerr != nil {
				l.Error(qerr.Error())
				return &core.LoadMyJoinedGroupResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
			}
			groups := make([]*core.GroupBriefItem, 0, len(groupList))
			for _, group := range groupList {
				groups = append(groups, &core.GroupBriefItem{
					GroupId:   group.Uuid,
					GroupName: group.Name,
					Avatar:    defaultAvatarIfEmpty(group.Avatar),
				})
			}
			if data, merr := json.Marshal(groups); merr != nil {
				l.Error(merr.Error())
			} else if serr := redisx.SetKeyEx(cacheKey, string(data), time.Minute*constants.REDIS_TIMEOUT); serr != nil {
				l.Error(serr.Error())
			}
			return &core.LoadMyJoinedGroupResp{Code: 200, Message: "获取成功", Groups: groups}, nil
		}
		l.Error(err.Error())
		return &core.LoadMyJoinedGroupResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	var groups []*core.GroupBriefItem
	if err := json.Unmarshal([]byte(rspString), &groups); err != nil {
		l.Error(err.Error())
	}
	return &core.LoadMyJoinedGroupResp{Code: 200, Message: "获取成功", Groups: groups}, nil
}
