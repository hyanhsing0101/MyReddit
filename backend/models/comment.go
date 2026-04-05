package models

import (
	"database/sql"
	"time"
)

type Comment struct {
	ID               int64          `db:"id"`
	PostID           int64          `db:"post_id"`
	AuthorID         sql.NullInt64  `db:"author_id"`
	ParentID         sql.NullInt64  `db:"parent_id"`
	Content          string         `db:"content"`
	DeletedAt        sql.NullTime   `db:"deleted_at"`
	CreateTime       time.Time      `db:"create_time"`
	UpdateTime       time.Time      `db:"update_time"`
	AuthorUsername   sql.NullString `db:"author_username"`
}

type CommentView struct {
	ID             int64     `json:"id"`
	PostID         int64     `json:"post_id"`
	AuthorID       *int64    `json:"author_id"`
	AuthorUsername string    `json:"author_username"`
	ParentID       *int64    `json:"parent_id"`
	Content        string    `json:"content"`
	CreateTime     time.Time `json:"create_time"`
	UpdateTime     time.Time `json:"update_time"`
}

func CommentToView(c Comment) CommentView {
	v := CommentView{
		ID:         c.ID,
		PostID:     c.PostID,
		Content:    c.Content,
		CreateTime: c.CreateTime,
		UpdateTime: c.UpdateTime,
	}
	if c.AuthorID.Valid {
		id := c.AuthorID.Int64
		v.AuthorID = &id
	}
	if c.AuthorUsername.Valid {
		v.AuthorUsername = c.AuthorUsername.String
	}
	if c.ParentID.Valid {
		pid := c.ParentID.Int64
		v.ParentID = &pid
	}
	return v
}

type CommentListData struct {
	List     []CommentView `json:"list"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}
