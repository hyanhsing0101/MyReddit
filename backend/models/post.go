package models

import (
	"database/sql"
	"time"
)

type Post struct {
	ID         int64          `db:"id"`
	BoardID    int64          `db:"board_id"`
	Title      string         `db:"title"`
	Content    string         `db:"content"`
	AuthorID   sql.NullInt64  `db:"author_id"`
	DeletedAt  sql.NullTime   `db:"deleted_at"`
	Score      int64          `db:"score"`
	CreateTime time.Time      `db:"create_time"`
	UpdateTime time.Time      `db:"update_time"`
	BoardSlug  sql.NullString `db:"board_slug"`
	BoardName  sql.NullString `db:"board_name"`
	TagIDs     []int64        `db:"tag_ids"`
}

// PostView 列表/详情接口返回用，author_id 可空时用 JSON null
type PostView struct {
	ID        int64  `json:"id"`
	BoardID   int64  `json:"board_id"`
	BoardSlug string `json:"board_slug"`
	BoardName string `json:"board_name"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	AuthorID  *int64 `json:"author_id"`
	Score     int64  `json:"score"`
	MyVote    *int8  `json:"my_vote"`
	// IsFavorited 仅当请求带合法登录态时设置 true/false；未登录时省略。
	IsFavorited *bool     `json:"is_favorited,omitempty"`
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
	Tags        []Tag     `json:"tags"`
}

func PostToView(p Post) PostView {
	v := PostView{
		ID:         p.ID,
		BoardID:    p.BoardID,
		Title:      p.Title,
		Content:    p.Content,
		Score:      p.Score,
		CreateTime: p.CreateTime,
		UpdateTime: p.UpdateTime,
	}
	if p.BoardSlug.Valid {
		v.BoardSlug = p.BoardSlug.String
	}
	if p.BoardName.Valid {
		v.BoardName = p.BoardName.String
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

// PostFavoriteView 我收藏的帖子：帖子字段 + 收藏时间
type PostFavoriteView struct {
	PostView
	FavoritedAt time.Time `json:"favorited_at"`
}

type PostFavoriteListData struct {
	List     []PostFavoriteView `json:"list"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
}

// PostVoteResult 投票接口返回：最新净分与当前用户投票状态（未投为 null）。
type PostVoteResult struct {
	Score  int64 `json:"score"`
	MyVote *int8 `json:"my_vote"`
}
