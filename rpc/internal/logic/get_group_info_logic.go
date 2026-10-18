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

type GetGroupInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetGroupInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupInfoLogic {
	return &GetGroupInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetGroupInfoLogic) GetGroupInfo(in *core.GetGroupInfoReq) (*core.GetGroupInfoResp, error) {
	groupId := in.GroupId
	rspString, err := redisx.GetKeyNilIsErr("group_info_" + groupId)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			group, qerr := l.svcCtx.GroupInfoModel.FindOneAllByUuid(l.ctx, groupId)
			if qerr != nil {
				l.Error(qerr.Error())
				return &core.GetGroupInfoResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
			}
			groupInfo := &core.GroupInfo{
				Uuid:      group.Uuid,
				Name:      group.Name,
				Notice:    group.Notice.String,
				MemberCnt: int32(group.MemberCnt),
				OwnerId:   group.OwnerId,
				AddMode:   int32(group.AddMode),
				Status:    int32(group.Status),
				Avatar:    defaultAvatarIfEmpty(group.Avatar),
				IsDeleted: group.DeletedAt.Valid,
			}
			return &core.GetGroupInfoResp{Code: 200, Message: "获取成功", Group: groupInfo}, nil
		}
		l.Error(err.Error())
		return &core.GetGroupInfoResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	var cached core.GroupInfo
	if err := json.Unmarshal([]byte(rspString), &cached); err != nil {
		l.Error(err.Error())
	}
	return &core.GetGroupInfoResp{Code: 200, Message: "获取成功", Group: &cached}, nil
}
