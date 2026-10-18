package model

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ SessionModel = (*customSessionModel)(nil)

type (
	// SessionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSessionModel.
	SessionModel interface {
		sessionModel
		withSession(session sqlx.Session) SessionModel

		// UpdateDeletedAtBySendOrReceive 软删除某用户作为发送或接收方的所有会话
		UpdateDeletedAtBySendOrReceive(ctx context.Context, deletedAt time.Time, uuid string) error
		// UpdateDeletedAtBySendReceive 软删除指定发送方和接收方的会话
		UpdateDeletedAtBySendReceive(ctx context.Context, deletedAt time.Time, sendId string, receiveId string) error
		// UpdateReceiveNameAvatarByReceiveId 根据接收方更新会话名称和头像
		UpdateReceiveNameAvatarByReceiveId(ctx context.Context, receiveName string, avatar string, receiveId string) error
		// DismissNowByReceiveId 按接收方软删除会话（NOW()）
		DismissNowByReceiveId(ctx context.Context, receiveId string) error
		// DismissNowBySendReceive 按发送方和接收方软删除会话（NOW()）
		DismissNowBySendReceive(ctx context.Context, sendId string, receiveId string) error
		// FindOneBySendReceive 查询指定发送方和接收方未删除的会话
		FindOneBySendReceive(ctx context.Context, sendId string, receiveId string) (*Session, error)
		// FindBySendId 查询发送方的会话列表
		FindBySendId(ctx context.Context, sendId string) ([]Session, error)
		// FindOneNotDeletedByUuid 根据uuid查询未删除的会话
		FindOneNotDeletedByUuid(ctx context.Context, uuid string) (*Session, error)
		// DismissNowByUuid 根据uuid软删除会话（NOW()）
		DismissNowByUuid(ctx context.Context, uuid string) error
	}

	customSessionModel struct {
		*defaultSessionModel
	}
)

// NewSessionModel returns a model for the database table.
func NewSessionModel(conn sqlx.SqlConn) SessionModel {
	return &customSessionModel{
		defaultSessionModel: newSessionModel(conn),
	}
}

func (m *customSessionModel) withSession(session sqlx.Session) SessionModel {
	return NewSessionModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customSessionModel) UpdateDeletedAtBySendOrReceive(ctx context.Context, deletedAt time.Time, uuid string) error {
	query := "UPDATE session SET deleted_at = ? WHERE (send_id = ? OR receive_id = ?) AND deleted_at IS NULL"
	_, err := m.conn.ExecCtx(ctx, query, deletedAt, uuid, uuid)
	return err
}

func (m *customSessionModel) UpdateDeletedAtBySendReceive(ctx context.Context, deletedAt time.Time, sendId string, receiveId string) error {
	query := "UPDATE session SET deleted_at = ? WHERE send_id = ? AND receive_id = ? AND deleted_at IS NULL"
	_, err := m.conn.ExecCtx(ctx, query, deletedAt, sendId, receiveId)
	return err
}

func (m *customSessionModel) UpdateReceiveNameAvatarByReceiveId(ctx context.Context, receiveName string, avatar string, receiveId string) error {
	query := "UPDATE `session` SET `receive_name` = ?, `avatar` = ? WHERE `receive_id` = ? AND `deleted_at` IS NULL"
	_, err := m.conn.ExecCtx(ctx, query, receiveName, avatar, receiveId)
	return err
}

func (m *customSessionModel) DismissNowByReceiveId(ctx context.Context, receiveId string) error {
	query := "UPDATE `session` SET `deleted_at` = NOW() WHERE `receive_id` = ? AND `deleted_at` IS NULL"
	_, err := m.conn.ExecCtx(ctx, query, receiveId)
	return err
}

func (m *customSessionModel) DismissNowBySendReceive(ctx context.Context, sendId string, receiveId string) error {
	query := "UPDATE `session` SET `deleted_at` = NOW() WHERE `send_id` = ? AND `receive_id` = ?"
	_, err := m.conn.ExecCtx(ctx, query, sendId, receiveId)
	return err
}

func (m *customSessionModel) FindOneBySendReceive(ctx context.Context, sendId string, receiveId string) (*Session, error) {
	var resp Session
	query := "SELECT * FROM `session` WHERE `send_id` = ? AND `receive_id` = ? AND `deleted_at` IS NULL LIMIT 1"
	err := m.conn.QueryRowCtx(ctx, &resp, query, sendId, receiveId)
	return &resp, err
}

func (m *customSessionModel) FindBySendId(ctx context.Context, sendId string) ([]Session, error) {
	var resp []Session
	query := "SELECT * FROM `session` WHERE `send_id` = ? AND `deleted_at` IS NULL ORDER BY `created_at` DESC"
	err := m.conn.QueryRowsCtx(ctx, &resp, query, sendId)
	return resp, err
}

func (m *customSessionModel) FindOneNotDeletedByUuid(ctx context.Context, uuid string) (*Session, error) {
	var resp Session
	query := "SELECT * FROM `session` WHERE `uuid` = ? AND `deleted_at` IS NULL LIMIT 1"
	err := m.conn.QueryRowCtx(ctx, &resp, query, uuid)
	return &resp, err
}

func (m *customSessionModel) DismissNowByUuid(ctx context.Context, uuid string) error {
	query := "UPDATE `session` SET `deleted_at` = NOW() WHERE `uuid` = ?"
	_, err := m.conn.ExecCtx(ctx, query, uuid)
	return err
}
