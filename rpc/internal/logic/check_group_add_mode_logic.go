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

type CheckGroupAddModeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckGroupAddModeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckGroupAddModeLogic {
	return &CheckGroupAddModeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CheckGroupAddModeLogic) CheckGroupAddMode(in *core.CheckGroupAddModeReq) (*core.CheckGroupAddModeResp, error) {
	groupId := in.GroupId
	rspString, err := redisx.GetKeyNilIsErr("group_info_" + groupId)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			group, qerr := l.svcCtx.GroupInfoModel.FindOneAllByUuid(l.ctx, groupId)
			if qerr != nil {
				l.Error(qerr.Error())
				return &core.CheckGroupAddModeResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
			}
			return &core.CheckGroupAddModeResp{Code: 200, Message: "加群方式获取成功", AddMode: int32(group.AddMode)}, nil
		}
		l.Error(err.Error())
		return &core.CheckGroupAddModeResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	var cached core.GroupInfo
	if err := json.Unmarshal([]byte(rspString), &cached); err != nil {
		l.Error(err.Error())
	}
	return &core.CheckGroupAddModeResp{Code: 200, Message: "加群方式获取成功", AddMode: cached.AddMode}, nil
}
