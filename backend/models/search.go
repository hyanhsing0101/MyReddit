package models

import "time"

type SearchPostItem struct {
	ID         int64     `db:"id" json:"id"`
	BoardID    int64     `db:"board_id" json:"board_id"`
	BoardSlug  string    `db:"board_slug" json:"board_slug"`
	BoardName  string    `db:"board_name" json:"board_name"`
	Title      string    `db:"title" json:"title"`
	Content    string    `db:"content" json:"content"`
	AuthorID   *int64    `db:"author_id" json:"author_id"`
	CreateTime time.Time `db:"create_time" json:"create_time"`
	UpdateTime time.Time `db:"update_time" json:"update_time"`
	Score      float64   `db:"score" json:"score"`
}

type SearchBoardItem struct {
	ID           int64     `db:"id" json:"id"`
	Slug         string    `db:"slug" json:"slug"`
	Name         string    `db:"name" json:"name"`
	Description  string    `db:"description" json:"description"`
	IsSystemSink bool      `db:"is_system_sink" json:"is_system_sink"`
	CreateTime   time.Time `db:"create_time" json:"create_time"`
	UpdateTime   time.Time `db:"update_time" json:"update_time"`
	Score        float64   `db:"score" json:"score"`
}

type SearchData struct {
	Query  string            `json:"query"`
	Scope  string            `json:"scope"`
	Posts  []SearchPostItem  `json:"posts"`
	Boards []SearchBoardItem `json:"boards"`
}