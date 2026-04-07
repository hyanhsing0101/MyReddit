package models

import (
	"database/sql"
	"time"
)

type Board struct {
	ID           int64          `db:"id"`
	Slug         string         `db:"slug"`
	Name         string         `db:"name"`
	Description  sql.NullString `db:"description"`
	CreatedBy    sql.NullInt64  `db:"created_by"`
	IsSystemSink bool           `db:"is_system_sink"`
	CreateTime   time.Time      `db:"create_time"`
	UpdateTime   time.Time      `db:"update_time"`
}

type BoardView struct {
	ID           int64     `json:"id"`
	Slug         string    `json:"slug"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	CreatedBy    *int64    `json:"created_by"`
	IsSystemSink bool      `json:"is_system_sink"`
	CreateTime   time.Time `json:"create_time"`
	UpdateTime   time.Time `json:"update_time"`
	// IsFavorited 仅当请求带合法登录态时设置 true/false；未登录时省略。
	IsFavorited *bool `json:"is_favorited,omitempty"`
}

func BoardToView(b Board) BoardView {
	v := BoardView{
		ID:           b.ID,
		Slug:         b.Slug,
		Name:         b.Name,
		IsSystemSink: b.IsSystemSink,
		CreateTime:   b.CreateTime,
		UpdateTime:   b.UpdateTime,
	}
	if b.Description.Valid {
		v.Description = b.Description.String
	}
	if b.CreatedBy.Valid {
		id := b.CreatedBy.Int64
		v.CreatedBy = &id
	}
	return v
}

type BoardListData struct {
	List     []BoardView `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// BoardFavoriteView 我收藏的板块：板块字段 + 收藏时间
type BoardFavoriteView struct {
	BoardView
	FavoritedAt time.Time `json:"favorited_at"`
}

type BoardFavoriteListData struct {
	List     []BoardFavoriteView `json:"list"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
}
