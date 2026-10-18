package model

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserContactModel = (*customUserContactModel)(nil)

type (
	// UserContactModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserContactModel.
	UserContactModel interface {
		userContactModel
		withSession(session sqlx.Session) UserContactModel

		// FindByUserIdExcludeStatus 根据用户id查询联系人列表，排除指定状态
		FindByUserIdExcludeStatus(ctx context.Context, userId string, status int64) ([]UserContact, error)
		// FindByUserIdExcludeTwoStatus 根据用户id查询联系人列表，排除两个指定状态
		FindByUserIdExcludeTwoStatus(ctx context.Context, userId string, status1 int64, status2 int64) ([]UserContact, error)
		// FindOneAllByUserAndContact 根据用户id和联系人id查询联系人全部字段
		FindOneAllByUserAndContact(ctx context.Context, userId string, contactId string) (*UserContact, error)
		// FindOneByUserAndContact 根据用户id和联系人id查询联系人状态相关字段
		FindOneByUserAndContact(ctx context.Context, userId string, contactId string) (*UserContact, error)
		// InsertUserContact 插入联系人记录
		InsertUserContact(ctx context.Context, userId string, contactId string, contactType int64, status int64, createdAt time.Time, updateAt time.Time) error
		// UpdateDeletedAtStatusByUserAndContact 更新联系人的删除时间和状态
		UpdateDeletedAtStatusByUserAndContact(ctx context.Context, deletedAt time.Time, status int64, userId string, contactId string) error
		// UpdateStatusUpdateAtByUserAndContact 更新联系人状态和更新时间
		UpdateStatusUpdateAtByUserAndContact(ctx context.Context, status int64, updateAt time.Time, userId string, contactId string) error
		// UpdateStatusByUserAndContact 更新联系人状态
		UpdateStatusByUserAndContact(ctx context.Context, status int64, userId string, contactId string) error
		// SoftDeleteByUserOrContact 软删除用户或联系人为指定值的所有联系人
		SoftDeleteByUserOrContact(ctx context.Context, deletedAt time.Time, userId string, contactId string) error
		// SoftDeleteNowByContactId 按联系人id软删除（NOW()）
		SoftDeleteNowByContactId(ctx context.Context, contactId string) error
		// SoftDeleteNowAndStatusByUserAndContact 软删除并更新联系人状态（NOW()）
		SoftDeleteNowAndStatusByUserAndContact(ctx context.Context, status int64, userId string, contactId string) error
		// SoftDeleteNowByUserAndContact 按用户id和联系人id软删除（NOW()）
		SoftDeleteNowByUserAndContact(ctx context.Context, userId string, contactId string) error
	}

	customUserContactModel struct {
		*defaultUserContactModel
	}
)

// NewUserContactModel returns a model for the database table.
func NewUserContactModel(conn sqlx.SqlConn) UserContactModel {
	return &customUserContactModel{
		defaultUserContactModel: newUserContactModel(conn),
	}
}

func (m *customUserContactModel) withSession(session sqlx.Session) UserContactModel {
	return NewUserContactModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customUserContactModel) FindByUserIdExcludeStatus(ctx context.Context, userId string, status int64) ([]UserContact, error) {
	var resp []UserContact
	query := "SELECT user_id, contact_id, contact_type, status FROM user_contact WHERE user_id = ? AND status != ? AND deleted_at IS NULL ORDER BY created_at DESC"
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId, status)
	return resp, err
}

func (m *customUserContactModel) FindByUserIdExcludeTwoStatus(ctx context.Context, userId string, status1 int64, status2 int64) ([]UserContact, error) {
	var resp []UserContact
	query := "SELECT user_id, contact_id, contact_type, status FROM user_contact WHERE user_id = ? AND status != ? AND status != ? AND deleted_at IS NULL ORDER BY created_at DESC"
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId, status1, status2)
	return resp, err
}

func (m *customUserContactModel) FindOneAllByUserAndContact(ctx context.Context, userId string, contactId string) (*UserContact, error) {
	var resp UserContact
	query := "SELECT * FROM `user_contact` WHERE `user_id` = ? AND `contact_id` = ? AND `deleted_at` IS NULL LIMIT 1"
	err := m.conn.QueryRowCtx(ctx, &resp, query, userId, contactId)
	return &resp, err
}

func (m *customUserContactModel) FindOneByUserAndContact(ctx context.Context, userId string, contactId string) (*UserContact, error) {
	var resp UserContact
	query := "SELECT id, user_id, contact_id, status FROM user_contact WHERE user_id = ? AND contact_id = ? AND deleted_at IS NULL LIMIT 1"
	err := m.conn.QueryRowCtx(ctx, &resp, query, userId, contactId)
	return &resp, err
}

func (m *customUserContactModel) InsertUserContact(ctx context.Context, userId string, contactId string, contactType int64, status int64, createdAt time.Time, updateAt time.Time) error {
	query := "INSERT INTO user_contact (user_id, contact_id, contact_type, status, created_at, update_at) VALUES (?, ?, ?, ?, ?, ?)"
	_, err := m.conn.ExecCtx(ctx, query, userId, contactId, contactType, status, createdAt, updateAt)
	return err
}

func (m *customUserContactModel) UpdateDeletedAtStatusByUserAndContact(ctx context.Context, deletedAt time.Time, status int64, userId string, contactId string) error {
	query := "UPDATE user_contact SET deleted_at = ?, status = ? WHERE user_id = ? AND contact_id = ? AND deleted_at IS NULL"
	_, err := m.conn.ExecCtx(ctx, query, deletedAt, status, userId, contactId)
	return err
}

func (m *customUserContactModel) UpdateStatusUpdateAtByUserAndContact(ctx context.Context, status int64, updateAt time.Time, userId string, contactId string) error {
	query := "UPDATE user_contact SET status = ?, update_at = ? WHERE user_id = ? AND contact_id = ? AND deleted_at IS NULL"
	_, err := m.conn.ExecCtx(ctx, query, status, updateAt, userId, contactId)
	return err
}

func (m *customUserContactModel) UpdateStatusByUserAndContact(ctx context.Context, status int64, userId string, contactId string) error {
	query := "UPDATE user_contact SET status = ? WHERE user_id = ? AND contact_id = ?"
	_, err := m.conn.ExecCtx(ctx, query, status, userId, contactId)
	return err
}

func (m *customUserContactModel) SoftDeleteByUserOrContact(ctx context.Context, deletedAt time.Time, userId string, contactId string) error {
	query := "UPDATE user_contact SET deleted_at = ? WHERE (user_id = ? OR contact_id = ?) AND deleted_at IS NULL"
	_, err := m.conn.ExecCtx(ctx, query, deletedAt, userId, contactId)
	return err
}

func (m *customUserContactModel) SoftDeleteNowByContactId(ctx context.Context, contactId string) error {
	query := "UPDATE `user_contact` SET `deleted_at` = NOW() WHERE `contact_id` = ? AND `deleted_at` IS NULL"
	_, err := m.conn.ExecCtx(ctx, query, contactId)
	return err
}

func (m *customUserContactModel) SoftDeleteNowAndStatusByUserAndContact(ctx context.Context, status int64, userId string, contactId string) error {
	query := "UPDATE `user_contact` SET `deleted_at` = NOW(), `status` = ? WHERE `user_id` = ? AND `contact_id` = ?"
	_, err := m.conn.ExecCtx(ctx, query, status, userId, contactId)
	return err
}

func (m *customUserContactModel) SoftDeleteNowByUserAndContact(ctx context.Context, userId string, contactId string) error {
	query := "UPDATE `user_contact` SET `deleted_at` = NOW() WHERE `user_id` = ? AND `contact_id` = ?"
	_, err := m.conn.ExecCtx(ctx, query, userId, contactId)
	return err
}
