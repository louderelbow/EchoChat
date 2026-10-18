package logic

import (
	"context"
	"database/sql"
	"encoding/json"

	"echochat/internal/model"
	"echochat/pkg/constants"
	"echochat/pkg/enum"
	"echochat/pkg/random"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateGroupLogic {
	return &CreateGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 群聊
func (l *CreateGroupLogic) CreateGroup(in *core.CreateGroupReq) (*core.BaseResp, error) {
	groupUuid := "G" + random.GetNowAndLenRandomString(11)
	members, err := json.Marshal([]string{in.OwnerId})
	if err != nil {
		l.Error(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	group := &model.GroupInfo{
		Uuid:      groupUuid,
		Name:      in.Name,
		Notice:    sql.NullString{String: in.Notice, Valid: true},
		Members:   sql.NullString{String: string(members), Valid: true},
		MemberCnt: 1,
		OwnerId:   in.OwnerId,
		AddMode:   int64(in.AddMode),
		Avatar:    defaultAvatarIfEmpty(in.Avatar),
		Status:    int64(enum.GroupNormal),
	}
	if _, err := l.svcCtx.GroupInfoModel.Insert(l.ctx, group); err != nil {
		l.Error(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	// 添加联系人
	contact := &model.UserContact{
		UserId:      in.OwnerId,
		ContactId:   groupUuid,
		ContactType: int64(enum.ContactGroup),
		Status:      int64(enum.ContactNormal),
	}
	if _, err := l.svcCtx.UserContactModel.Insert(l.ctx, contact); err != nil {
		l.Error(err.Error())
		return &core.BaseResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}

	return &core.BaseResp{Code: 200, Message: "创建成功"}, nil
}
