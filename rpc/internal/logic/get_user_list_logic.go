package logic

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	redis "github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"

	"echochat/pkg/constants"
	"echochat/pkg/enum"
	"echochat/pkg/redisx"
	"echochat/rpc/core"
	"echochat/rpc/internal/svc"
)

type GetUserListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserListLogic {
	return &GetUserListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 联系人
func (l *GetUserListLogic) GetUserList(in *core.OwnlistReq) (*core.MyUserListResp, error) {
	rspString, err := redisx.GetKeyNilIsErr("contact_user_list_" + in.OwnerId)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			contactList, e := l.svcCtx.UserContactModel.FindByUserIdExcludeStatus(l.ctx, in.OwnerId, int64(enum.ContactDelete))
			if e != nil {
				l.Errorf(e.Error())
				return &core.MyUserListResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
			}

			var userListRsp []*core.MyUserListItem
			for _, contact := range contactList {
				if contact.ContactType == int64(enum.ContactUser) {
					user, e := l.svcCtx.UserInfoModel.FindOneBriefByUuid(l.ctx, contact.ContactId)
					if e != nil {
						l.Errorf(e.Error())
						return &core.MyUserListResp{Code: 500, Message: constants.SYSTEM_ERROR}, nil
					}
					userListRsp = append(userListRsp, &core.MyUserListItem{
						UserId:   user.Uuid,
						UserName: user.Nickname,
						Avatar:   user.Avatar,
					})
				}
			}

			if b, e := json.Marshal(userListRsp); e != nil {
				l.Errorf(e.Error())
			} else if e := redisx.SetKeyEx("contact_user_list_"+in.OwnerId, string(b), time.Minute*time.Duration(constants.REDIS_TIMEOUT)); e != nil {
				l.Errorf(e.Error())
			}
			return &core.MyUserListResp{Code: 200, Message: "获取用户列表成功", Users: userListRsp}, nil
		}
		l.Errorf(err.Error())
	}

	var rsp []*core.MyUserListItem
	if err := json.Unmarshal([]byte(rspString), &rsp); err != nil {
		l.Errorf(err.Error())
	}
	return &core.MyUserListResp{Code: 200, Message: "获取用户列表成功", Users: rsp}, nil
}
