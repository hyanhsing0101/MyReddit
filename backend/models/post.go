package models

import (
	"database/sql"
	"time"
)

type Post struct {
	ID         int64         `db:"id"`
	Title      string        `db:"title"`
	Content    string        `db:"content"`
	AuthorID   sql.NullInt64 `db:"author_id"`
	CreateTime time.Time     `db:"create_time"`
	UpdateTime time.Time     `db:"update_time"`
}

// PostView 列表/详情接口返回用，author_id 可空时用 JSON null
type PostView struct {
	ID         int64     `json:"id"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	AuthorID   *int64    `json:"author_id"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

func PostToView(p Post) PostView {
	v := PostView{
		ID:         p.ID,
		Title:      p.Title,
		Content:    p.Content,
		CreateTime: p.CreateTime,
		UpdateTime: p.UpdateTime,
	}
	if p.AuthorID.Valid {
		id := p.AuthorID.Int64
		v.AuthorID = &id
	}
	return v
}

type PostListData struct {
	List     []PostView `json:"list"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}
