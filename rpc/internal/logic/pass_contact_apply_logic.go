package logic

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"echochat/pkg/constants"
	"echochat/pkg/enum"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"
)

type PassContactApplyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPassContactApplyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PassContactApplyLogic {
	return &PassContactApplyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PassContactApplyLogic) PassContactApply(in *core.ContactPairReq) (*core.BaseResp, error) {
	// ownerId 如果是用户就是登录用户，如果是群聊就是群聊id
	apply, err := l.svcCtx.ContactApplyModel.FindOneStatusByContactUser(l.ctx, in.OwnerId, in.ContactId)
	if err != nil {
		l.Errorf(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	now := time.Now()

	if len(in.OwnerId) > 0 && in.OwnerId[0] == 'U' {
		user, e := l.svcCtx.UserInfoModel.FindOneStatusByUuid(l.ctx, in.ContactId)
		if e != nil && !errors.Is(e, sqlx.ErrNotFound) {
			l.Errorf(e.Error())
		}
		if user.Status == int64(enum.UserDisable) {
			return &core.BaseResp{Code: 400, Message: "用户已被禁用"}, nil
		}
		if e := l.svcCtx.ContactApplyModel.UpdateStatusByUuid(l.ctx, int64(enum.ApplyAgree), apply.Uuid); e != nil {
			l.Errorf(e.Error())
			return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
		}
		if e := l.svcCtx.UserContactModel.InsertUserContact(l.ctx, in.OwnerId, in.ContactId, int64(enum.ContactUser), int64(enum.ContactNormal), now, now); e != nil {
			l.Errorf(e.Error())
			return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
		}
		if e := l.svcCtx.UserContactModel.InsertUserContact(l.ctx, in.ContactId, in.OwnerId, int64(enum.ContactUser), int64(enum.ContactNormal), now, now); e != nil {
			l.Errorf(e.Error())
			return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
		}
		return &core.BaseResp{Code: 200, Message: "已添加该联系人"}, nil
	}

	// 群聊
	group, e := l.svcCtx.GroupInfoModel.FindOneMembersByUuid(l.ctx, in.OwnerId)
	if e != nil {
		l.Errorf(e.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	if group.Status == int64(enum.GroupDisable) {
		return &core.BaseResp{Code: 400, Message: "群聊已被禁用"}, nil
	}
	if e := l.svcCtx.ContactApplyModel.UpdateStatusByUuid(l.ctx, int64(enum.ApplyAgree), apply.Uuid); e != nil {
		l.Errorf(e.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	// 群聊只创建一个 UserContact，因为一个 UserContact 足以表达双方的状态
	if e := l.svcCtx.UserContactModel.InsertUserContact(l.ctx, in.ContactId, in.OwnerId, int64(enum.ContactGroup), int64(enum.ContactNormal), now, now); e != nil {
		l.Errorf(e.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	var members []string
	if e := json.Unmarshal([]byte(group.Members.String), &members); e != nil {
		l.Errorf(e.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	members = append(members, in.ContactId)
	memberBytes, _ := json.Marshal(members)
	if e := l.svcCtx.GroupInfoModel.UpdateMembersAndCntByUuid(l.ctx, string(memberBytes), int64(len(members)), time.Now(), in.OwnerId); e != nil {
		l.Errorf(e.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	return &core.BaseResp{Code: 200, Message: "已通过加群申请"}, nil
}
