package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserModel = (*customUserModel)(nil)

type (
	// userModel 基础模型接口
	userModel interface {
		Insert(ctx context.Context, session sqlx.Session, data interface{}) (sql.Result, error)
		FindOne(ctx context.Context, session sqlx.Session, id int64) (interface{}, error)
		Update(ctx context.Context, session sqlx.Session, data interface{}) error
		Delete(ctx context.Context, session sqlx.Session, id int64) error
		tableName() string
	}

	// UserModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserModel.
	UserModel interface {
		userModel
		// 只添加额外的方法，不要重复基础方法
		FindOneByEmail(ctx context.Context, email string) (*User, error)
	}

	customUserModel struct {
		*defaultUserModel
	}

	// User represents the user data structure
	User struct {
		Id        int64     `db:"id"`
		Name      string    `db:"name"`
		Email     string    `db:"email"`
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}
)

// NewUserModel returns a model for the model for the database table.
func NewUserModel(conn sqlx.SqlConn) UserModel {
	return &customUserModel{
		defaultUserModel: newUserModel(conn),
	}
}

// 实现基础接口方法（通过defaultUserModel）

func (m *customUserModel) FindOneByEmail(ctx context.Context, email string) (*User, error) {
	query := `SELECT id, name, email, created_at, updated_at FROM users WHERE email = ?`
	var user User
	err := m.conn.QueryRowCtx(ctx, &user, query, email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}