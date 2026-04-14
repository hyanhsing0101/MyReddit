package models

import "time"

type ModerationAction string

const (
	ModerationActionSealPost             ModerationAction = "seal_post"
	ModerationActionUnsealPost           ModerationAction = "unseal_post"
	ModerationActionDeletePost           ModerationAction = "delete_post"
	ModerationActionRestorePost          ModerationAction = "restore_post"
	ModerationActionHandlePostAppeal     ModerationAction = "handle_post_appeal"
	ModerationActionLockPostComments     ModerationAction = "lock_post_comments"
	ModerationActionUnlockPostComments   ModerationAction = "unlock_post_comments"
	ModerationActionPinPost              ModerationAction = "pin_post"
	ModerationActionUnpinPost            ModerationAction = "unpin_post"
	ModerationActionResolvePostReport    ModerationAction = "update_post_report_status"
	ModerationActionResolveCommentReport ModerationAction = "update_comment_report_status"
	ModerationActionUpsertBoardModerator ModerationAction = "upsert_board_moderator"
	ModerationActionUpdateBoardModerator ModerationAction = "update_board_moderator_role"
	ModerationActionRemoveBoardModerator ModerationAction = "remove_board_moderator"
)

type ModerationTargetType string

const (
	ModerationTargetPost          ModerationTargetType = "post"
	ModerationTargetPostReport    ModerationTargetType = "post_report"
	ModerationTargetCommentReport ModerationTargetType = "comment_report"
	ModerationTargetModerator     ModerationTargetType = "board_moderator"
)

type ModerationLog struct {
	ID          int64     `db:"id"`
	BoardID     int64     `db:"board_id"`
	OperatorID  int64     `db:"operator_id"`
	Action      string    `db:"action"`
	TargetType  string    `db:"target_type"`
	TargetID    int64     `db:"target_id"`
	Description string    `db:"description"`
	CreateTime  time.Time `db:"create_time"`
}

type ModerationLogView struct {
	ID               int64     `json:"id" db:"id"`
	BoardID          int64     `json:"board_id" db:"board_id"`
	OperatorID       int64     `json:"operator_id" db:"operator_id"`
	OperatorUsername string    `json:"operator_username" db:"operator_username"`
	Action           string    `json:"action" db:"action"`
	TargetType       string    `json:"target_type" db:"target_type"`
	TargetID         int64     `json:"target_id" db:"target_id"`
	Description      string    `json:"description" db:"description"`
	CreateTime       time.Time `json:"create_time" db:"create_time"`
}

type ModerationLogListData struct {
	List     []ModerationLogView `json:"list"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
}
