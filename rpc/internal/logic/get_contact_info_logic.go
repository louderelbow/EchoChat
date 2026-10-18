package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"echochat/pkg/constants"
	"echochat/pkg/enum"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"
)

type GetContactInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetContactInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetContactInfoLogic {
	return &GetContactInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetContactInfoLogic) GetContactInfo(in *core.GetContactInfoReq) (*core.GetContactInfoResp, error) {
	if len(in.ContactId) > 0 && in.ContactId[0] == 'G' {
		group, err := l.svcCtx.GroupInfoModel.FindOneContactByUuid(l.ctx, in.ContactId)
		if err != nil {
			l.Errorf(err.Error())
			return &core.GetContactInfoResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
		}
		if group.Status != int64(enum.GroupDisable) {
			return &core.GetContactInfoResp{
				Code:    200,
				Message: "获取联系人信息成功",
				Contact: &core.ContactInfo{
					ContactId:        group.Uuid,
					ContactName:      group.Name,
					ContactAvatar:    defaultAvatarIfEmpty(group.Avatar),
					ContactNotice:    group.Notice.String,
					ContactAddMode:   int32(group.AddMode),
					ContactMembers:   group.Members.String,
					ContactMemberCnt: int32(group.MemberCnt),
					ContactOwnerId:   group.OwnerId,
				},
			}, nil
		}
		return &core.GetContactInfoResp{Code: 400, Message: "该群聊处于禁用状态"}, nil
	}

	user, err := l.svcCtx.UserInfoModel.FindOneContactByUuid(l.ctx, in.ContactId)
	if err != nil {
		l.Errorf(err.Error())
		return &core.GetContactInfoResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
	}
	if user.Status != int64(enum.UserDisable) {
		return &core.GetContactInfoResp{
			Code:    200,
			Message: "获取联系人信息成功",
			Contact: &core.ContactInfo{
				ContactId:        user.Uuid,
				ContactName:      user.Nickname,
				ContactAvatar:    user.Avatar,
				ContactBirthday:  user.Birthday.String,
				ContactEmail:     user.Email.String,
				ContactPhone:     user.Telephone,
				ContactGender:    int32(user.Gender),
				ContactSignature: user.Signature.String,
			},
		}, nil
	}
	return &core.GetContactInfoResp{Code: 400, Message: "该用户处于禁用状态"}, nil
}
