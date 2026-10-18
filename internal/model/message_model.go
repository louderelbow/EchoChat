package model

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ MessageModel = (*customMessageModel)(nil)

type (
	// MessageModel is an interface to be customized, add more methods here,
	// and implement the added methods in customMessageModel.
	MessageModel interface {
		messageModel
		withSession(session sqlx.Session) MessageModel

		// FindConversationByUsers 查询两个用户之间的聊天记录
		FindConversationByUsers(ctx context.Context, userOneId string, userTwoId string) ([]Message, error)
		// FindByReceiveId 根据接收方id查询聊天记录
		FindByReceiveId(ctx context.Context, receiveId string) ([]Message, error)
	}

	customMessageModel struct {
		*defaultMessageModel
	}
)

// NewMessageModel returns a model for the database table.
func NewMessageModel(conn sqlx.SqlConn) MessageModel {
	return &customMessageModel{
		defaultMessageModel: newMessageModel(conn),
	}
}

func (m *customMessageModel) withSession(session sqlx.Session) MessageModel {
	return NewMessageModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customMessageModel) FindConversationByUsers(ctx context.Context, userOneId string, userTwoId string) ([]Message, error) {
	var resp []Message
	query := "SELECT * FROM `message` WHERE (`send_id` = ? AND `receive_id` = ?) OR (`send_id` = ? AND `receive_id` = ?) ORDER BY `created_at` ASC"
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userOneId, userTwoId, userTwoId, userOneId)
	return resp, err
}

func (m *customMessageModel) FindByReceiveId(ctx context.Context, receiveId string) ([]Message, error) {
	var resp []Message
	query := "SELECT * FROM `message` WHERE `receive_id` = ? ORDER BY `created_at` ASC"
	err := m.conn.QueryRowsCtx(ctx, &resp, query, receiveId)
	return resp, err
}
