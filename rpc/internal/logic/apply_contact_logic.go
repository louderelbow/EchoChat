package logic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"echochat/internal/model"
	"echochat/pkg/constants"
	"echochat/pkg/enum"
	"echochat/pkg/random"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"
)

type ApplyContactLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewApplyContactLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApplyContactLogic {
	return &ApplyContactLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ApplyContactLogic) ApplyContact(in *core.ApplyContactReq) (*core.BaseResp, error) {
	var contactType int8
	switch {
	case len(in.ContactId) > 0 && in.ContactId[0] == 'U':
		contactType = enum.ContactUser
		user, err := l.svcCtx.UserInfoModel.FindOneStatusByUuid(l.ctx, in.ContactId)
		if err != nil {
			if errors.Is(err, sqlx.ErrNotFound) {
				return &core.BaseResp{Code: 400, Message: "用户不存在"}, nil
			}
			l.Errorf(err.Error())
			return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
		}
		if user.Status == int64(enum.UserDisable) {
			return &core.BaseResp{Code: 400, Message: "用户已被禁用"}, nil
		}
	case len(in.ContactId) > 0 && in.ContactId[0] == 'G':
		contactType = enum.ContactGroup
		group, err := l.svcCtx.GroupInfoModel.FindOneStatusByUuid(l.ctx, in.ContactId)
		if err != nil {
			if errors.Is(err, sqlx.ErrNotFound) {
				return &core.BaseResp{Code: 400, Message: "群聊不存在"}, nil
			}
			l.Errorf(err.Error())
			return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
		}
		if group.Status == int64(enum.GroupDisable) {
			return &core.BaseResp{Code: 400, Message: "群聊已被禁用"}, nil
		}
	default:
		return &core.BaseResp{Code: 400, Message: "用户/群聊不存在"}, nil
	}

	apply, err := l.svcCtx.ContactApplyModel.FindOneApplyByUserContact(l.ctx, in.OwnerId, in.ContactId)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			apply = &model.ContactApply{
				Uuid:        "A" + random.GetNowAndLenRandomString(11),
				UserId:      in.OwnerId,
				ContactId:   in.ContactId,
				ContactType: int64(contactType),
				Status:      int64(enum.ApplyPending),
				Message:     sql.NullString{String: in.Message, Valid: true},
				LastApplyAt: time.Now(),
			}
			if _, e := l.svcCtx.ContactApplyModel.Insert(l.ctx, apply); e != nil {
				l.Errorf(e.Error())
				return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
			}
		} else {
			l.Errorf(err.Error())
			return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
		}
	}

	if contactType == enum.ContactUser {
		// 存在申请记录，先看看有没有被拉黑
		if apply.Status == int64(enum.ApplyBlack) {
			return &core.BaseResp{Code: 400, Message: "对方已将你拉黑"}, nil
		}
		if e := l.svcCtx.ContactApplyModel.UpdateLastApplyAtStatusByUuid(l.ctx, time.Now(), int64(enum.ApplyPending), apply.Uuid); e != nil {
			l.Errorf(e.Error())
			return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
		}
	} else {
		if e := l.svcCtx.ContactApplyModel.UpdateLastApplyAtByUuid(l.ctx, time.Now(), apply.Uuid); e != nil {
			l.Errorf(e.Error())
			return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
		}
	}
	return &core.BaseResp{Code: 200, Message: "申请成功"}, nil
}
