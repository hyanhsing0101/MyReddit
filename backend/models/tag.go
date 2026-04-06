package models

import "time"

type Tag struct {
	ID          int64     `db:"id" json:"id"`
	Slug        string    `db:"slug" json:"slug"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	CreateTime  time.Time `db:"create_time" json:"create_time"`
	UpdateTime  time.Time `db:"update_time" json:"update_time"`
}

type PostTag struct {
	PostID     int64     `db:"post_id" json:"post_id"`
	TagID      int64     `db:"tag_id" json:"tag_id"`
	CreateTime time.Time `db:"create_time" json:"create_time"`
}

type TagListData struct {
	List     []Tag `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}
