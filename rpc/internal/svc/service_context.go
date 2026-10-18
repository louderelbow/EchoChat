package svc

import (
	"echochat/internal/model"
	"echochat/pkg/db"
	"echochat/pkg/es"
	"echochat/pkg/mq"
	"echochat/pkg/redisx"
	"echochat/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config config.Config
	Conn   sqlx.SqlConn
	Mq     *mq.Client

	UserInfoModel     model.UserInfoModel
	GroupInfoModel    model.GroupInfoModel
	UserContactModel  model.UserContactModel
	SessionModel      model.SessionModel
	ContactApplyModel model.ContactApplyModel
	MessageModel      model.MessageModel
	Es                *es.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := db.NewMysql(c.Mysql)
	redisx.Init(redisx.Conf{Host: c.RedisConf.Host, Port: c.RedisConf.Port, Password: c.RedisConf.Password, Db: c.RedisConf.Db})
	// 验证码邮件服务通过环境变量配置（ECHO_MAIL_*），见 pkg/mail
	m, err := mq.NewClient(c.RabbitMQ, "")
	if err != nil {
		panic(err)
	}
	return &ServiceContext{
		Config:            c,
		Conn:              conn,
		Mq:                m,
		UserInfoModel:     model.NewUserInfoModel(conn),
		GroupInfoModel:    model.NewGroupInfoModel(conn),
		UserContactModel:  model.NewUserContactModel(conn),
		SessionModel:      model.NewSessionModel(conn),
		ContactApplyModel: model.NewContactApplyModel(conn),
		MessageModel:      model.NewMessageModel(conn),
		Es:                es.New(c.Elastic),
	}
}
