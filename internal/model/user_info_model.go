package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserInfoModel = (*customUserInfoModel)(nil)

type (
	// UserInfoModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserInfoModel.
	UserInfoModel interface {
		userInfoModel
		withSession(session sqlx.Session) UserInfoModel

		// FindOneByTelephone 根据手机号查询用户（含密码），用于账号密码登录
		FindOneByTelephone(ctx context.Context, telephone string) (*UserInfo, error)
		// FindOneByTelephoneNoPassword 根据手机号查询用户（不含密码），用于验证码登录
		FindOneByTelephoneNoPassword(ctx context.Context, telephone string) (*UserInfo, error)
		// FindIdByTelephone 根据手机号查询用户id，用于注册时判断手机号是否已存在
		FindIdByTelephone(ctx context.Context, telephone string) (*UserInfo, error)
		// FindOneEditableByUuid 查询用户可编辑字段，用于更新用户信息前取值
		FindOneEditableByUuid(ctx context.Context, uuid string) (*UserInfo, error)
		// FindOneProfileByUuid 根据uuid查询用户资料，用于获取用户信息
		FindOneProfileByUuid(ctx context.Context, uuid string) (*UserInfo, error)
		// FindOneStatusByUuid 根据uuid查询用户状态
		FindOneStatusByUuid(ctx context.Context, uuid string) (*UserInfo, error)
		// FindOneBriefByUuid 根据uuid查询用户简要信息（uuid、昵称、头像）
		FindOneBriefByUuid(ctx context.Context, uuid string) (*UserInfo, error)
		// FindOneContactByUuid 根据uuid查询用户联系人信息
		FindOneContactByUuid(ctx context.Context, uuid string) (*UserInfo, error)
		// FindOneAllByUuid 根据uuid查询用户全部字段
		FindOneAllByUuid(ctx context.Context, uuid string) (*UserInfo, error)
		// UpdateUserInfo 更新用户邮箱、昵称、性别、生日、签名、头像
		UpdateUserInfo(ctx context.Context, email sql.NullString, nickname string, gender int64, birthday sql.NullString, signature sql.NullString, avatar string, uuid string) error
		// UpdatePasswordByUuid 根据uuid更新用户密码哈希（用于明文密码的平滑升级）
		UpdatePasswordByUuid(ctx context.Context, password string, uuid string) error
		// UpdateStatusByUuid 根据uuid更新用户状态
		UpdateStatusByUuid(ctx context.Context, status int64, uuid string) error
		// SoftDeleteByUuid 根据uuid软删除用户
		SoftDeleteByUuid(ctx context.Context, deletedAt time.Time, uuid string) error
		// UpdateAdminByUuid 根据uuid更新用户管理员标记
		UpdateAdminByUuid(ctx context.Context, isAdmin int64, uuid string) error
		// FindAllExceptUuid 查询除指定uuid外的所有用户（含已软删除）
		FindAllExceptUuid(ctx context.Context, uuid string) ([]UserInfo, error)
		// InsertRegisterUser 注册时插入新用户
		InsertRegisterUser(ctx context.Context, uuid, nickname, telephone, email, avatar string, gender int64, signature, password, birthday string, isAdmin int64, status int64, createdAt time.Time) error
	}

	customUserInfoModel struct {
		*defaultUserInfoModel
	}
)

// NewUserInfoModel returns a model for the database table.
func NewUserInfoModel(conn sqlx.SqlConn) UserInfoModel {
	return &customUserInfoModel{
		defaultUserInfoModel: newUserInfoModel(conn),
	}
}

func (m *customUserInfoModel) withSession(session sqlx.Session) UserInfoModel {
	return NewUserInfoModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customUserInfoModel) FindOneByTelephone(ctx context.Context, telephone string) (*UserInfo, error) {
	var resp UserInfo
	query := "SELECT uuid, nickname, telephone, email, avatar, gender, birthday, signature, password, created_at, is_admin, status FROM user_info WHERE telephone = ? AND deleted_at IS NULL LIMIT 1"
	err := m.conn.QueryRowCtx(ctx, &resp, query, telephone)
	return &resp, err
}

func (m *customUserInfoModel) FindOneByTelephoneNoPassword(ctx context.Context, telephone string) (*UserInfo, error) {
	var resp UserInfo
	query := "SELECT uuid, nickname, telephone, email, avatar, gender, birthday, signature, created_at, is_admin, status FROM user_info WHERE telephone = ? AND deleted_at IS NULL LIMIT 1"
	err := m.conn.QueryRowCtx(ctx, &resp, query, telephone)
	return &resp, err
}

func (m *customUserInfoModel) FindIdByTelephone(ctx context.Context, telephone string) (*UserInfo, error) {
	var resp UserInfo
	query := "SELECT id FROM user_info WHERE telephone = ? AND deleted_at IS NULL LIMIT 1"
	err := m.conn.QueryRowCtx(ctx, &resp, query, telephone)
	return &resp, err
}

func (m *customUserInfoModel) FindOneEditableByUuid(ctx context.Context, uuid string) (*UserInfo, error) {
	var resp UserInfo
	query := "SELECT id, uuid, nickname, email, avatar, gender, birthday, signature FROM user_info WHERE uuid = ? AND deleted_at IS NULL LIMIT 1"
	err := m.conn.QueryRowCtx(ctx, &resp, query, uuid)
	return &resp, err
}

func (m *customUserInfoModel) FindOneProfileByUuid(ctx context.Context, uuid string) (*UserInfo, error) {
	var resp UserInfo
	query := "SELECT uuid, nickname, telephone, email, avatar, gender, birthday, signature, created_at, is_admin, status FROM user_info WHERE uuid = ? AND deleted_at IS NULL LIMIT 1"
	err := m.conn.QueryRowCtx(ctx, &resp, query, uuid)
	return &resp, err
}

func (m *customUserInfoModel) FindOneStatusByUuid(ctx context.Context, uuid string) (*UserInfo, error) {
	var resp UserInfo
	query := "SELECT uuid, status FROM user_info WHERE uuid = ? AND deleted_at IS NULL LIMIT 1"
	err := m.conn.QueryRowCtx(ctx, &resp, query, uuid)
	return &resp, err
}

func (m *customUserInfoModel) FindOneBriefByUuid(ctx context.Context, uuid string) (*UserInfo, error) {
	var resp UserInfo
	query := "SELECT uuid, nickname, avatar FROM user_info WHERE uuid = ? AND deleted_at IS NULL LIMIT 1"
	err := m.conn.QueryRowCtx(ctx, &resp, query, uuid)
	return &resp, err
}

func (m *customUserInfoModel) FindOneContactByUuid(ctx context.Context, uuid string) (*UserInfo, error) {
	var resp UserInfo
	query := "SELECT uuid, nickname, avatar, birthday, email, gender, telephone, signature, status FROM user_info WHERE uuid = ? AND deleted_at IS NULL LIMIT 1"
	err := m.conn.QueryRowCtx(ctx, &resp, query, uuid)
	return &resp, err
}

func (m *customUserInfoModel) FindOneAllByUuid(ctx context.Context, uuid string) (*UserInfo, error) {
	var resp UserInfo
	query := "SELECT * FROM `user_info` WHERE `uuid` = ? AND `deleted_at` IS NULL LIMIT 1"
	err := m.conn.QueryRowCtx(ctx, &resp, query, uuid)
	return &resp, err
}

func (m *customUserInfoModel) UpdateUserInfo(ctx context.Context, email sql.NullString, nickname string, gender int64, birthday sql.NullString, signature sql.NullString, avatar string, uuid string) error {
	query := "UPDATE user_info SET email = ?, nickname = ?, gender = ?, birthday = ?, signature = ?, avatar = ? WHERE uuid = ?"
	_, err := m.conn.ExecCtx(ctx, query, email, nickname, gender, birthday, signature, avatar, uuid)
	return err
}

func (m *customUserInfoModel) UpdatePasswordByUuid(ctx context.Context, password string, uuid string) error {
	query := "UPDATE user_info SET password = ? WHERE uuid = ?"
	_, err := m.conn.ExecCtx(ctx, query, password, uuid)
	return err
}

func (m *customUserInfoModel) UpdateStatusByUuid(ctx context.Context, status int64, uuid string) error {
	query := "UPDATE user_info SET status = ? WHERE uuid = ? AND deleted_at IS NULL"
	_, err := m.conn.ExecCtx(ctx, query, status, uuid)
	return err
}

func (m *customUserInfoModel) SoftDeleteByUuid(ctx context.Context, deletedAt time.Time, uuid string) error {
	query := "UPDATE user_info SET deleted_at = ? WHERE uuid = ? AND deleted_at IS NULL"
	_, err := m.conn.ExecCtx(ctx, query, deletedAt, uuid)
	return err
}

func (m *customUserInfoModel) UpdateAdminByUuid(ctx context.Context, isAdmin int64, uuid string) error {
	query := "UPDATE user_info SET is_admin = ? WHERE uuid = ? AND deleted_at IS NULL"
	_, err := m.conn.ExecCtx(ctx, query, isAdmin, uuid)
	return err
}

func (m *customUserInfoModel) FindAllExceptUuid(ctx context.Context, uuid string) ([]UserInfo, error) {
	var resp []UserInfo
	query := "SELECT uuid, nickname, telephone, status, is_admin, deleted_at FROM user_info WHERE uuid != ?"
	err := m.conn.QueryRowsCtx(ctx, &resp, query, uuid)
	return resp, err
}

func (m *customUserInfoModel) InsertRegisterUser(ctx context.Context, uuid, nickname, telephone, email, avatar string, gender int64, signature, password, birthday string, isAdmin int64, status int64, createdAt time.Time) error {
	query := "INSERT INTO user_info (uuid, nickname, telephone, email, avatar, gender, signature, password, birthday, is_admin, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	_, err := m.conn.ExecCtx(ctx, query, uuid, nickname, telephone, email, avatar, gender, signature, password, birthday, isAdmin, status, createdAt)
	return err
}
