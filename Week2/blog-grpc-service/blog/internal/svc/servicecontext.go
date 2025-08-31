package svc

import (
	"blog/blog/internal/config"
	"blog/blog/internal/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config    config.Config
	BlogModel model.BlogModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewSqlConn("sqlite3", c.DataSource)
	return &ServiceContext{
		Config:    c,
		BlogModel: model.NewBlogModel(conn),
	}
}
