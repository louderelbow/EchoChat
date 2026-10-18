package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ GroupInfoModel = (*customGroupInfoModel)(nil)

type (
	// GroupInfoModel is an interface to be customized, add more methods here,
	// and implement the added methods in customGroupInfoModel.
	GroupInfoModel interface {
		groupInfoModel
		withSession(session sqlx.Session) GroupInfoModel

		// FindOneContactByUuid 根据uuid查询群聊联系人信息
		FindOneContactByUuid(ctx context.Context, uuid string) (*GroupInfo, error)
		// FindOneStatusByUuid 根据uuid查询群聊状态
		FindOneStatusByUuid(ctx context.Context, uuid string) (*GroupInfo, error)
		// FindOneMembersByUuid 根据uuid查询群聊成员相关字段
		FindOneMembersByUuid(ctx context.Context, uuid string) (*GroupInfo, error)
		// UpdateMembersAndCntByUuid 更新群聊成员列表和人数，并更新updated_at
		UpdateMembersAndCntByUuid(ctx context.Context, members string, memberCnt int64, updatedAt time.Time, uuid string) error
		// FindOneBriefByUuid 根据uuid查询群聊简要信息（uuid、名称、头像、群主）
		FindOneBriefByUuid(ctx context.Context, uuid string) (*GroupInfo, error)
		// FindOneAllByUuid 根据uuid查询群聊全部字段
		FindOneAllByUuid(ctx context.Context, uuid string) (*GroupInfo, error)
		// UpdateMembersAndCntNoTimeByUuid 更新群聊成员列表和人数（不更新时间）
		UpdateMembersAndCntNoTimeByUuid(ctx context.Context, members string, memberCnt int64, uuid string) error
		// UpdateNameAddModeNoticeAvatarByUuid 更新群聊名称、加群方式、公告、头像
		UpdateNameAddModeNoticeAvatarByUuid(ctx context.Context, name string, addMode int64, notice sql.NullString, avatar string, uuid string) error
		// DismissByUuid 解散群聊（软删除并更新updated_at）
		DismissByUuid(ctx context.Context, uuid string) error
		// SoftDeleteByUuid 软删除群聊
		SoftDeleteByUuid(ctx context.Context, uuid string) error
		// UpdateStatusByUuid 根据uuid更新群聊状态
		UpdateStatusByUuid(ctx context.Context, status int64, uuid string) error
		// FindAll 查询所有群聊（含已软删除）
		FindAll(ctx context.Context) ([]GroupInfo, error)
		// FindByOwnerId 根据群主uuid查询群聊列表
		FindByOwnerId(ctx context.Context, ownerId string) ([]GroupInfo, error)
	}

	customGroupInfoModel struct {
		*defaultGroupInfoModel
	}
)

// NewGroupInfoModel returns a model for the database table.
func NewGroupInfoModel(conn sqlx.SqlConn) GroupInfoModel {
	return &customGroupInfoModel{
		defaultGroupInfoModel: newGroupInfoModel(conn),
	}
}

func (m *customGroupInfoModel) withSession(session sqlx.Session) GroupInfoModel {
	return NewGroupInfoModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customGroupInfoModel) FindOneContactByUuid(ctx context.Context, uuid string) (*GroupInfo, error) {
	var resp GroupInfo
	query := "SELECT uuid, name, notice, members, member_cnt, owner_id, add_mode, status FROM group_info WHERE uuid = ? AND deleted_at IS NULL LIMIT 1"
	err := m.conn.QueryRowCtx(ctx, &resp, query, uuid)
	return &resp, err
}

func (m *customGroupInfoModel) FindOneStatusByUuid(ctx context.Context, uuid string) (*GroupInfo, error) {
	var resp GroupInfo
	query := "SELECT uuid, status FROM group_info WHERE uuid = ? AND deleted_at IS NULL LIMIT 1"
	err := m.conn.QueryRowCtx(ctx, &resp, query, uuid)
	return &resp, err
}

func (m *customGroupInfoModel) FindOneMembersByUuid(ctx context.Context, uuid string) (*GroupInfo, error) {
	var resp GroupInfo
	query := "SELECT uuid, members, member_cnt, status FROM group_info WHERE uuid = ? AND deleted_at IS NULL LIMIT 1"
	err := m.conn.QueryRowCtx(ctx, &resp, query, uuid)
	return &resp, err
}

func (m *customGroupInfoModel) UpdateMembersAndCntByUuid(ctx context.Context, members string, memberCnt int64, updatedAt time.Time, uuid string) error {
	query := "UPDATE group_info SET members = ?, member_cnt = ?, updated_at = ? WHERE uuid = ?"
	_, err := m.conn.ExecCtx(ctx, query, members, memberCnt, updatedAt, uuid)
	return err
}

func (m *customGroupInfoModel) FindOneBriefByUuid(ctx context.Context, uuid string) (*GroupInfo, error) {
	var resp GroupInfo
	query := "SELECT uuid, name, avatar, owner_id FROM group_info WHERE uuid = ? AND deleted_at IS NULL LIMIT 1"
	err := m.conn.QueryRowCtx(ctx, &resp, query, uuid)
	return &resp, err
}

func (m *customGroupInfoModel) FindOneAllByUuid(ctx context.Context, uuid string) (*GroupInfo, error) {
	var resp GroupInfo
	query := "SELECT * FROM `group_info` WHERE `uuid` = ? AND `deleted_at` IS NULL LIMIT 1"
	err := m.conn.QueryRowCtx(ctx, &resp, query, uuid)
	return &resp, err
}

func (m *customGroupInfoModel) UpdateMembersAndCntNoTimeByUuid(ctx context.Context, members string, memberCnt int64, uuid string) error {
	query := "UPDATE `group_info` SET `members` = ?, `member_cnt` = ? WHERE `uuid` = ?"
	_, err := m.conn.ExecCtx(ctx, query, members, memberCnt, uuid)
	return err
}

func (m *customGroupInfoModel) UpdateNameAddModeNoticeAvatarByUuid(ctx context.Context, name string, addMode int64, notice sql.NullString, avatar string, uuid string) error {
	query := "UPDATE `group_info` SET `name` = ?, `add_mode` = ?, `notice` = ?, `avatar` = ? WHERE `uuid` = ?"
	_, err := m.conn.ExecCtx(ctx, query, name, addMode, notice, avatar, uuid)
	return err
}

func (m *customGroupInfoModel) DismissByUuid(ctx context.Context, uuid string) error {
	query := "UPDATE `group_info` SET `deleted_at` = NOW(), `updated_at` = NOW() WHERE `uuid` = ?"
	_, err := m.conn.ExecCtx(ctx, query, uuid)
	return err
}

func (m *customGroupInfoModel) SoftDeleteByUuid(ctx context.Context, uuid string) error {
	query := "UPDATE `group_info` SET `deleted_at` = NOW() WHERE `uuid` = ?"
	_, err := m.conn.ExecCtx(ctx, query, uuid)
	return err
}

func (m *customGroupInfoModel) UpdateStatusByUuid(ctx context.Context, status int64, uuid string) error {
	query := "UPDATE `group_info` SET `status` = ? WHERE `uuid` = ?"
	_, err := m.conn.ExecCtx(ctx, query, status, uuid)
	return err
}

func (m *customGroupInfoModel) FindAll(ctx context.Context) ([]GroupInfo, error) {
	var resp []GroupInfo
	query := "SELECT * FROM `group_info`"
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	return resp, err
}

func (m *customGroupInfoModel) FindByOwnerId(ctx context.Context, ownerId string) ([]GroupInfo, error) {
	var resp []GroupInfo
	query := "SELECT * FROM `group_info` WHERE `owner_id` = ? AND `deleted_at` IS NULL ORDER BY `created_at` DESC"
	err := m.conn.QueryRowsCtx(ctx, &resp, query, ownerId)
	return resp, err
}
