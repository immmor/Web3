package model

import (
	"context"
	"database/sql"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type defaultUserModel struct {
	conn sqlx.SqlConn
}

func newUserModel(conn sqlx.SqlConn) *defaultUserModel {
	return &defaultUserModel{conn: conn}
}

func (m *defaultUserModel) tableName() string {
	return "users"
}

// 实现基础接口方法
func (m *defaultUserModel) Insert(ctx context.Context, session sqlx.Session, data interface{}) (sql.Result, error) {
	// 基础实现，可以根据需要完善
	return nil, nil
}

func (m *defaultUserModel) FindOne(ctx context.Context, session sqlx.Session, id int64) (interface{}, error) {
	// 基础实现，可以根据需要完善
	return nil, nil
}

func (m *defaultUserModel) Update(ctx context.Context, session sqlx.Session, data interface{}) error {
	// 基础实现，可以根据需要完善
	return nil
}

func (m *defaultUserModel) Delete(ctx context.Context, session sqlx.Session, id int64) error {
	// 基础实现，可以根据需要完善
	return nil
}