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
