package db

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type MysqlConf struct {
	Host         string
	Port         int
	User         string
	Password     string
	DatabaseName string
}

// DB 包装 go-zero 的 SqlConn。
//
// go-zero 的 QueryRowCtx / QueryRowsCtx 是「严格模式」：要求查询列数 ≥ 目标结构体字段数，
// 否则在查到行时返回 ErrNotMatchDestination。本项目大量使用「部分列查询 + 完整 model 结构体」，
// 因此这里把这两个方法覆盖为「非严格」实现（unmarshalRow(..., false)），使未查询到的字段保持零值。
type DB struct {
	sqlx.SqlConn
}

func NewMysql(conf MysqlConf) *DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		conf.User, conf.Password, conf.Host, conf.Port, conf.DatabaseName)
	return &DB{SqlConn: sqlx.NewMysql(dsn)}
}

// QueryRowCtx 覆盖为非严格版本：允许目标结构体字段多于查询列数
func (d *DB) QueryRowCtx(ctx context.Context, v any, query string, args ...any) error {
	return d.SqlConn.QueryRowPartialCtx(ctx, v, query, args...)
}

// QueryRowsCtx 覆盖为非严格版本：允许目标结构体字段多于查询列数
func (d *DB) QueryRowsCtx(ctx context.Context, v any, query string, args ...any) error {
	return d.SqlConn.QueryRowsPartialCtx(ctx, v, query, args...)
}
