package model

import (
	"context"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ContactApplyModel = (*customContactApplyModel)(nil)

type (
	// ContactApplyModel is an interface to be customized, add more methods here,
	// and implement the added methods in customContactApplyModel.
	ContactApplyModel interface {
		contactApplyModel
		withSession(session sqlx.Session) ContactApplyModel

		// FindOneStatusByContactUser 根据被申请人id和申请人id查询申请状态
		FindOneStatusByContactUser(ctx context.Context, contactId string, userId string) (*ContactApply, error)
		// FindOneApplyByUserContact 根据申请人id和被申请人id查询申请记录
		FindOneApplyByUserContact(ctx context.Context, userId string, contactId string) (*ContactApply, error)
		// FindOneBasicByContactUser 根据被申请人id和申请人id查询申请基础信息
		FindOneBasicByContactUser(ctx context.Context, contactId string, userId string) (*ContactApply, error)
		// FindByContactIdAndStatus 根据被申请人id和状态查询申请列表
		FindByContactIdAndStatus(ctx context.Context, contactId string, status int64) ([]ContactApply, error)
		// CountByContactIdAndStatus 统计指定被申请人/群聊下某状态的申请数量
		CountByContactIdAndStatus(ctx context.Context, contactId string, status int64) (int64, error)
		// CountByContactIdsAndStatus 统计多个被申请人/群聊下某状态的申请数量之和
		CountByContactIdsAndStatus(ctx context.Context, contactIds []string, status int64) (int64, error)
		// UpdateStatusByUuid 根据uuid更新申请状态
		UpdateStatusByUuid(ctx context.Context, status int64, uuid string) error
		// UpdateLastApplyAtStatusByUuid 根据uuid更新最后申请时间和状态
		UpdateLastApplyAtStatusByUuid(ctx context.Context, lastApplyAt time.Time, status int64, uuid string) error
		// UpdateLastApplyAtByUuid 根据uuid更新最后申请时间
		UpdateLastApplyAtByUuid(ctx context.Context, lastApplyAt time.Time, uuid string) error
		// SoftDeleteByUserOrContact 软删除申请人或被申请人为指定值的所有申请
		SoftDeleteByUserOrContact(ctx context.Context, deletedAt time.Time, userId string, contactId string) error
		// SoftDeleteByContactUser 按被申请人id和申请人id软删除申请
		SoftDeleteByContactUser(ctx context.Context, deletedAt time.Time, contactId string, userId string) error
		// SoftDeleteNowByContactId 按被申请人id软删除申请（NOW()）
		SoftDeleteNowByContactId(ctx context.Context, contactId string) error
		// SoftDeleteNowByContactUser 按被申请人id和申请人id软删除申请（NOW()）
		SoftDeleteNowByContactUser(ctx context.Context, contactId string, userId string) error
		// SoftDeleteNowByUserContact 按申请人id和被申请人id软删除申请（NOW()）
		SoftDeleteNowByUserContact(ctx context.Context, userId string, contactId string) error
	}

	customContactApplyModel struct {
		*defaultContactApplyModel
	}
)

// NewContactApplyModel returns a model for the database table.
func NewContactApplyModel(conn sqlx.SqlConn) ContactApplyModel {
	return &customContactApplyModel{
		defaultContactApplyModel: newContactApplyModel(conn),
	}
}

func (m *customContactApplyModel) withSession(session sqlx.Session) ContactApplyModel {
	return NewContactApplyModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customContactApplyModel) FindOneStatusByContactUser(ctx context.Context, contactId string, userId string) (*ContactApply, error) {
	var resp ContactApply
	query := "SELECT id, uuid, user_id, contact_id, status FROM contact_apply WHERE contact_id = ? AND user_id = ? AND deleted_at IS NULL LIMIT 1"
	err := m.conn.QueryRowCtx(ctx, &resp, query, contactId, userId)
	return &resp, err
}

func (m *customContactApplyModel) FindOneApplyByUserContact(ctx context.Context, userId string, contactId string) (*ContactApply, error) {
	var resp ContactApply
	query := "SELECT id, uuid, user_id, contact_id, contact_type, status, message, last_apply_at FROM contact_apply WHERE user_id = ? AND contact_id = ? AND deleted_at IS NULL LIMIT 1"
	err := m.conn.QueryRowCtx(ctx, &resp, query, userId, contactId)
	return &resp, err
}

func (m *customContactApplyModel) FindOneBasicByContactUser(ctx context.Context, contactId string, userId string) (*ContactApply, error) {
	var resp ContactApply
	query := "SELECT id, uuid FROM contact_apply WHERE contact_id = ? AND user_id = ? AND deleted_at IS NULL LIMIT 1"
	err := m.conn.QueryRowCtx(ctx, &resp, query, contactId, userId)
	return &resp, err
}

func (m *customContactApplyModel) FindByContactIdAndStatus(ctx context.Context, contactId string, status int64) ([]ContactApply, error) {
	var resp []ContactApply
	query := "SELECT id, uuid, user_id, contact_id, status, message FROM contact_apply WHERE contact_id = ? AND status = ? AND deleted_at IS NULL"
	err := m.conn.QueryRowsCtx(ctx, &resp, query, contactId, status)
	return resp, err
}

// CountByContactIdAndStatus 统计指定被申请人/群聊下某状态的申请数量。
func (m *customContactApplyModel) CountByContactIdAndStatus(ctx context.Context, contactId string, status int64) (int64, error) {
	var count int64
	query := "SELECT COUNT(*) FROM contact_apply WHERE contact_id = ? AND status = ? AND deleted_at IS NULL"
	err := m.conn.QueryRowCtx(ctx, &count, query, contactId, status)
	return count, err
}

// CountByContactIdsAndStatus 统计多个被申请人/群聊下某状态的申请数量之和。
func (m *customContactApplyModel) CountByContactIdsAndStatus(ctx context.Context, contactIds []string, status int64) (int64, error) {
	if len(contactIds) == 0 {
		return 0, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(contactIds)), ",")
	query := "SELECT COUNT(*) FROM contact_apply WHERE contact_id IN (" + placeholders + ") AND status = ? AND deleted_at IS NULL"
	args := make([]any, 0, len(contactIds)+1)
	for _, id := range contactIds {
		args = append(args, id)
	}
	args = append(args, status)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, args...)
	return count, err
}

func (m *customContactApplyModel) UpdateStatusByUuid(ctx context.Context, status int64, uuid string) error {
	query := "UPDATE contact_apply SET status = ? WHERE uuid = ?"
	_, err := m.conn.ExecCtx(ctx, query, status, uuid)
	return err
}

func (m *customContactApplyModel) UpdateLastApplyAtStatusByUuid(ctx context.Context, lastApplyAt time.Time, status int64, uuid string) error {
	query := "UPDATE contact_apply SET last_apply_at = ?, status = ? WHERE uuid = ?"
	_, err := m.conn.ExecCtx(ctx, query, lastApplyAt, status, uuid)
	return err
}

func (m *customContactApplyModel) UpdateLastApplyAtByUuid(ctx context.Context, lastApplyAt time.Time, uuid string) error {
	query := "UPDATE contact_apply SET last_apply_at = ? WHERE uuid = ?"
	_, err := m.conn.ExecCtx(ctx, query, lastApplyAt, uuid)
	return err
}

func (m *customContactApplyModel) SoftDeleteByUserOrContact(ctx context.Context, deletedAt time.Time, userId string, contactId string) error {
	query := "UPDATE contact_apply SET deleted_at = ? WHERE (user_id = ? OR contact_id = ?) AND deleted_at IS NULL"
	_, err := m.conn.ExecCtx(ctx, query, deletedAt, userId, contactId)
	return err
}

func (m *customContactApplyModel) SoftDeleteByContactUser(ctx context.Context, deletedAt time.Time, contactId string, userId string) error {
	query := "UPDATE contact_apply SET deleted_at = ? WHERE contact_id = ? AND user_id = ? AND deleted_at IS NULL"
	_, err := m.conn.ExecCtx(ctx, query, deletedAt, contactId, userId)
	return err
}

func (m *customContactApplyModel) SoftDeleteNowByContactId(ctx context.Context, contactId string) error {
	query := "UPDATE `contact_apply` SET `deleted_at` = NOW() WHERE `contact_id` = ? AND `deleted_at` IS NULL"
	_, err := m.conn.ExecCtx(ctx, query, contactId)
	return err
}

func (m *customContactApplyModel) SoftDeleteNowByContactUser(ctx context.Context, contactId string, userId string) error {
	query := "UPDATE `contact_apply` SET `deleted_at` = NOW() WHERE `contact_id` = ? AND `user_id` = ?"
	_, err := m.conn.ExecCtx(ctx, query, contactId, userId)
	return err
}

func (m *customContactApplyModel) SoftDeleteNowByUserContact(ctx context.Context, userId string, contactId string) error {
	query := "UPDATE `contact_apply` SET `deleted_at` = NOW() WHERE `user_id` = ? AND `contact_id` = ?"
	_, err := m.conn.ExecCtx(ctx, query, userId, contactId)
	return err
}
