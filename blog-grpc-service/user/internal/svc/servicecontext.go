package svc

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"blog/user/internal/config"  // 修改为相对于模块根的路径
	"blog/user/internal/model"    // 修改为相对于模块根的路径
)

type ServiceContext struct {
	Config    config.Config
	UserModel model.UserModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewSqlConn("sqlite3", c.DataSource)
	return &ServiceContext{
		Config:    c,
		UserModel: model.NewUserModel(conn),
	}
}
