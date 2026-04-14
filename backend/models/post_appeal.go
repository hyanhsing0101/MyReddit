package models

import "time"

type PostAppealStatus string

const (
	PostAppealStatusOpen     PostAppealStatus = "open"
	PostAppealStatusReview   PostAppealStatus = "in_review"
	PostAppealStatusApproved PostAppealStatus = "approved"
	PostAppealStatusRejected PostAppealStatus = "rejected"
)

type PostAppeal struct {
	ID             int64            `db:"id"`
	PostID         int64            `db:"post_id"`
	BoardID        int64            `db:"board_id"`
	AuthorID       int64            `db:"author_id"`
	Reason         string           `db:"reason"`
	RequestedTitle string           `db:"requested_title"`
	RequestedBody  string           `db:"requested_content"`
	UserReply      string           `db:"user_reply"`
	Status         PostAppealStatus `db:"status"`
	ModeratorID    *int64           `db:"moderator_id"`
	ModeratorReply string           `db:"moderator_reply"`
	CreateTime     time.Time        `db:"create_time"`
	UpdateTime     time.Time        `db:"update_time"`
}

type PostAppealView struct {
	ID             int64            `json:"id" db:"id"`
	PostID         int64            `json:"post_id" db:"post_id"`
	PostTitle      string           `json:"post_title" db:"post_title"`
	BoardID        int64            `json:"board_id" db:"board_id"`
	AuthorID       int64            `json:"author_id" db:"author_id"`
	AuthorUsername string           `json:"author_username" db:"author_username"`
	Reason         string           `json:"reason" db:"reason"`
	RequestedTitle string           `json:"requested_title" db:"requested_title"`
	RequestedBody  string           `json:"requested_content" db:"requested_content"`
	UserReply      string           `json:"user_reply" db:"user_reply"`
	Status         PostAppealStatus `json:"status" db:"status"`
	ModeratorID    *int64           `json:"moderator_id" db:"moderator_id"`
	ModeratorName  string           `json:"moderator_username" db:"moderator_username"`
	ModeratorReply string           `json:"moderator_reply" db:"moderator_reply"`
	CreateTime     time.Time        `json:"create_time" db:"create_time"`
	UpdateTime     time.Time        `json:"update_time" db:"update_time"`
}

type PostAppealListData struct {
	List     []PostAppealView `json:"list"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}
