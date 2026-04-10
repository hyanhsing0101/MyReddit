package models

import (
	"database/sql"
	"time"
)

const (
	BoardVisibilityPublic       = "public"
	BoardVisibilityPrivate      = "private"
	BoardModeratorRoleOwner     = "owner"
	BoardModeratorRoleModerator = "moderator"
)

type Board struct {
	ID           int64          `db:"id"`
	Slug         string         `db:"slug"`
	Name         string         `db:"name"`
	Description  sql.NullString `db:"description"`
	CreatedBy    sql.NullInt64  `db:"created_by"`
	Visibility   string         `db:"visibility"`
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
	Visibility   string    `json:"visibility"`
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
		Visibility:   b.Visibility,
		IsSystemSink: b.IsSystemSink,
		CreateTime:   b.CreateTime,
		UpdateTime:   b.UpdateTime,
	}
	if v.Visibility == "" {
		v.Visibility = BoardVisibilityPublic
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

// BoardModerator 板块版主关系（含 owner/moderator 角色）。
type BoardModerator struct {
	UserID      int64          `db:"user_id" json:"user_id"`
	BoardID     int64          `db:"board_id" json:"board_id"`
	Role        string         `db:"role" json:"role"`
	Username    sql.NullString `db:"username" json:"-"`
	AppointedBy sql.NullInt64  `db:"appointed_by" json:"appointed_by"`
	CreateTime  time.Time      `db:"create_time" json:"create_time"`
	UpdateTime  time.Time      `db:"update_time" json:"update_time"`
}

type BoardModeratorView struct {
	UserID      int64     `json:"user_id"`
	Username    string    `json:"username"`
	Role        string    `json:"role"`
	AppointedBy *int64    `json:"appointed_by"`
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
}

func BoardModeratorToView(m BoardModerator) BoardModeratorView {
	v := BoardModeratorView{
		UserID:     m.UserID,
		Role:       m.Role,
		CreateTime: m.CreateTime,
		UpdateTime: m.UpdateTime,
	}
	if m.Username.Valid {
		v.Username = m.Username.String
	}
	if m.AppointedBy.Valid {
		id := m.AppointedBy.Int64
		v.AppointedBy = &id
	}
	return v
}
