package models

import "time"

type User struct {
	UserID      int64  `db:"user_id"`
	Username    string `db:"username"`
	Password    string `db:"password"`
	IsSiteAdmin bool   `db:"is_site_admin"`
}

type MePermissionsView struct {
	UserID            int64   `json:"user_id"`
	Username          string  `json:"username"`
	Roles             []string `json:"roles"`
	IsSiteAdmin       bool    `json:"is_site_admin"`
	ModeratedBoardIDs []int64 `json:"moderated_board_ids"`
}

type UserHomePostItem struct {
	ID         int64     `db:"id" json:"id"`
	BoardID    int64     `db:"board_id" json:"board_id"`
	BoardSlug  string    `db:"board_slug" json:"board_slug"`
	BoardName  string    `db:"board_name" json:"board_name"`
	Title      string    `db:"title" json:"title"`
	Score      int64     `db:"score" json:"score"`
	CreateTime time.Time `db:"create_time" json:"create_time"`
	UpdateTime time.Time `db:"update_time" json:"update_time"`
}

type UserHomeCommentItem struct {
	ID         int64     `db:"id" json:"id"`
	PostID     int64     `db:"post_id" json:"post_id"`
	PostTitle  string    `db:"post_title" json:"post_title"`
	Content    string    `db:"content" json:"content"`
	Score      int64     `db:"score" json:"score"`
	CreateTime time.Time `db:"create_time" json:"create_time"`
	UpdateTime time.Time `db:"update_time" json:"update_time"`
}

type UserHomeData struct {
	UserID           int64                 `json:"user_id"`
	Username         string                `json:"username"`
	Posts            []UserHomePostItem    `json:"posts"`
	PostsTotal       int64                 `json:"posts_total"`
	PostPage         int                   `json:"post_page"`
	PostPageSize     int                   `json:"post_page_size"`
	Comments         []UserHomeCommentItem `json:"comments"`
	CommentsTotal    int64                 `json:"comments_total"`
	CommentPage      int                   `json:"comment_page"`
	CommentPageSize  int                   `json:"comment_page_size"`
}
