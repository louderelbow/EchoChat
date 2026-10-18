package logic

import (
	"context"
	"encoding/json"

	"echochat/internal/model"
	"echochat/pkg/constants"
	"echochat/pkg/enum"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type EnterGroupDirectlyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewEnterGroupDirectlyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EnterGroupDirectlyLogic {
	return &EnterGroupDirectlyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *EnterGroupDirectlyLogic) EnterGroupDirectly(in *core.ContactPairReq) (*core.BaseResp, error) {
	groupId := in.OwnerId
	contactId := in.ContactId

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
	members = append(members, contactId)
	data, err := json.Marshal(members)
	if err != nil {
		l.Error(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	if err := l.svcCtx.GroupInfoModel.UpdateMembersAndCntNoTimeByUuid(l.ctx, string(data), group.MemberCnt+1, groupId); err != nil {
		l.Error(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	// 添加联系人
	newContact := &model.UserContact{
		UserId:      contactId,
		ContactId:   groupId,
		ContactType: int64(enum.ContactGroup),
		Status:      int64(enum.ContactNormal),
	}
	if _, err := l.svcCtx.UserContactModel.Insert(l.ctx, newContact); err != nil {
		l.Error(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	return &core.BaseResp{Code: 200, Message: "进群成功"}, nil
}
