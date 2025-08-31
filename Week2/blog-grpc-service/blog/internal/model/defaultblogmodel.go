package model

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type defaultBlogModel struct {
	conn sqlx.SqlConn
}

func newBlogModel(conn sqlx.SqlConn) *defaultBlogModel {
	return &defaultBlogModel{conn: conn}
}

func (m *defaultBlogModel) tableName() string {
	return "blogs"
}