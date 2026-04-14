package models

import "time"

// CommentReportStatus 与帖子举报一致。
type CommentReportStatus string

const (
	CommentReportStatusOpen     CommentReportStatus = "open"
	CommentReportStatusReview   CommentReportStatus = "in_review"
	CommentReportStatusResolved CommentReportStatus = "resolved"
	CommentReportStatusRejected CommentReportStatus = "rejected"
)

type CommentReport struct {
	ID          int64               `db:"id"`
	CommentID   int64               `db:"comment_id"`
	PostID      int64               `db:"post_id"`
	BoardID     int64               `db:"board_id"`
	ReporterID  int64               `db:"reporter_id"`
	Reason      string              `db:"reason"`
	Detail      string              `db:"detail"`
	Status      CommentReportStatus `db:"status"`
	HandlerID   *int64              `db:"handler_id"`
	HandlerNote string              `db:"handler_note"`
	CreateTime  time.Time           `db:"create_time"`
	UpdateTime  time.Time           `db:"update_time"`
}

type CommentReportView struct {
	ID               int64               `json:"id" db:"id"`
	CommentID        int64               `json:"comment_id" db:"comment_id"`
	PostID           int64               `json:"post_id" db:"post_id"`
	PostTitle        string              `json:"post_title" db:"post_title"`
	CommentSnippet   string              `json:"comment_snippet" db:"comment_snippet"`
	BoardID          int64               `json:"board_id" db:"board_id"`
	ReporterID       int64               `json:"reporter_id" db:"reporter_id"`
	ReporterUsername string              `json:"reporter_username" db:"reporter_username"`
	Reason           string              `json:"reason" db:"reason"`
	Detail           string              `json:"detail" db:"detail"`
	Status           CommentReportStatus `json:"status" db:"status"`
	HandlerID        *int64              `json:"handler_id" db:"handler_id"`
	HandlerUsername  string              `json:"handler_username" db:"handler_username"`
	HandlerNote      string              `json:"handler_note" db:"handler_note"`
	CreateTime       time.Time           `json:"create_time" db:"create_time"`
	UpdateTime       time.Time           `json:"update_time" db:"update_time"`
}

type CommentReportListData struct {
	List        []CommentReportView `json:"list"`
	Total       int64               `json:"total"`
	PendingOpen int64               `json:"pending_open"`
	Page        int                 `json:"page"`
	PageSize    int                 `json:"page_size"`
}
