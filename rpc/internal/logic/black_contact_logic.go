package logic

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"echochat/pkg/constants"
	"echochat/pkg/enum"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"
)

type BlackContactLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBlackContactLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BlackContactLogic {
	return &BlackContactLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *BlackContactLogic) BlackContact(in *core.ContactPairReq) (*core.BaseResp, error) {
	now := time.Now()
	// 拉黑
	if err := l.svcCtx.UserContactModel.UpdateStatusUpdateAtByUserAndContact(l.ctx, int64(enum.ContactBlack), now, in.OwnerId, in.ContactId); err != nil {
		l.Errorf(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	// 被拉黑
	if err := l.svcCtx.UserContactModel.UpdateStatusUpdateAtByUserAndContact(l.ctx, int64(enum.ContactBeBlack), now, in.ContactId, in.OwnerId); err != nil {
		l.Errorf(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	// 删除会话
	if err := l.svcCtx.SessionModel.UpdateDeletedAtBySendReceive(l.ctx, now, in.OwnerId, in.ContactId); err != nil {
		l.Errorf(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	return &core.BaseResp{Code: 200, Message: "已拉黑该联系人"}, nil
}
