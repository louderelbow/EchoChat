package logic

import (
	"context"
	"errors"

	"echochat/pkg/constants"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type DeleteSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteSessionLogic {
	return &DeleteSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteSessionLogic) DeleteSession(in *core.DeleteSessionReq) (*core.BaseResp, error) {
	_, err := l.svcCtx.SessionModel.FindOneNotDeletedByUuid(l.ctx, in.SessionId)
	if err != nil && !errors.Is(err, sqlx.ErrNotFound) {
		l.Error(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	if err := l.svcCtx.SessionModel.DismissNowByUuid(l.ctx, in.SessionId); err != nil {
		l.Error(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	return &core.BaseResp{Code: 200, Message: "删除成功"}, nil
}
