package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ BlogModel = (*customBlogModel)(nil)

type (
	// BlogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customBlogModel.
	BlogModel interface {
		blogModel
		Insert(ctx context.Context, data *Blog) (sql.Result, error)
		FindOne(ctx context.Context, id int64) (*Blog, error)
		FindByAuthor(ctx context.Context, authorId int64) ([]*Blog, error)
		FindAll(ctx context.Context) ([]*Blog, error)
		Update(ctx context.Context, data *Blog) error
		Delete(ctx context.Context, id int64) error
	}

	customBlogModel struct {
		*defaultBlogModel
	}

	// Blog represents the blog data structure
	Blog struct {
		Id        int64     `db:"id"`
		Title     string    `db:"title"`
		Content   string    `db:"content"`
		AuthorId  int64     `db:"author_id"`
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}
)

// NewBlogModel returns a model for the database table.
func NewBlogModel(conn sqlx.SqlConn) BlogModel {
	return &customBlogModel{
		defaultBlogModel: newBlogModel(conn),
	}
}

func (m *customBlogModel) Insert(ctx context.Context, data *Blog) (sql.Result, error) {
	query := `INSERT INTO blogs (title, content, author_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`
	return m.conn.ExecCtx(ctx, query, data.Title, data.Content, data.AuthorId, time.Now(), time.Now())
}

func (m *customBlogModel) FindOne(ctx context.Context, id int64) (*Blog, error) {
	query := `SELECT id, title, content, author_id, created_at, updated_at FROM blogs WHERE id = ?`
	var blog Blog
	err := m.conn.QueryRowCtx(ctx, &blog, query, id)
	if err != nil {
		return nil, err
	}
	return &blog, nil
}

func (m *customBlogModel) FindByAuthor(ctx context.Context, authorId int64) ([]*Blog, error) {
	query := `SELECT id, title, content, author_id, created_at, updated_at FROM blogs WHERE author_id = ? ORDER BY created_at DESC`
	var blogs []*Blog
	err := m.conn.QueryRowsCtx(ctx, &blogs, query, authorId)
	if err != nil {
		return nil, err
	}
	return blogs, nil
}

func (m *customBlogModel) FindAll(ctx context.Context) ([]*Blog, error) {
	query := `SELECT id, title, content, author_id, created_at, updated_at FROM blogs ORDER BY created_at DESC`
	var blogs []*Blog
	err := m.conn.QueryRowsCtx(ctx, &blogs, query)
	if err != nil {
		return nil, err
	}
	return blogs, nil
}

func (m *customBlogModel) Update(ctx context.Context, data *Blog) error {
	query := `UPDATE blogs SET title = ?, content = ?, author_id = ?, updated_at = ? WHERE id = ?`
	_, err := m.conn.ExecCtx(ctx, query, data.Title, data.Content, data.AuthorId, time.Now(), data.Id)
	return err
}

func (m *customBlogModel) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM blogs WHERE id = ?`
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}