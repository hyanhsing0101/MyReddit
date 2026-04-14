package models

import "time"

type PostReportStatus string

const (
	PostReportStatusOpen     PostReportStatus = "open"
	PostReportStatusReview   PostReportStatus = "in_review"
	PostReportStatusResolved PostReportStatus = "resolved"
	PostReportStatusRejected PostReportStatus = "rejected"
)

type PostReport struct {
	ID          int64            `db:"id"`
	PostID      int64            `db:"post_id"`
	BoardID     int64            `db:"board_id"`
	ReporterID  int64            `db:"reporter_id"`
	Reason      string           `db:"reason"`
	Detail      string           `db:"detail"`
	Status      PostReportStatus `db:"status"`
	HandlerID   *int64           `db:"handler_id"`
	HandlerNote string           `db:"handler_note"`
	CreateTime  time.Time        `db:"create_time"`
	UpdateTime  time.Time        `db:"update_time"`
}

type PostReportView struct {
	ID               int64            `json:"id"`
	PostID           int64            `json:"post_id"`
	PostTitle        string           `json:"post_title"`
	BoardID          int64            `json:"board_id"`
	ReporterID       int64            `json:"reporter_id"`
	ReporterUsername string           `json:"reporter_username"`
	Reason           string           `json:"reason"`
	Detail           string           `json:"detail"`
	Status           PostReportStatus `json:"status"`
	HandlerID        *int64           `json:"handler_id"`
	HandlerUsername  string           `json:"handler_username"`
	HandlerNote      string           `json:"handler_note"`
	CreateTime       time.Time        `json:"create_time"`
	UpdateTime       time.Time        `json:"update_time"`
}

type PostReportListData struct {
	List        []PostReportView `json:"list"`
	Total       int64            `json:"total"`
	PendingOpen int64            `json:"pending_open"`
	Page        int              `json:"page"`
	PageSize    int              `json:"page_size"`
}
