package postgres

import (
	"time"
)

// InsertBoardModerator 写入版主行（建板事务内调用）。
func InsertBoardModerator(userID, boardID int64, at time.Time) error {
	_, err := db.Exec(`
		INSERT INTO board_moderator (user_id, board_id, create_time)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, board_id) DO NOTHING`,
		userID, boardID, at,
	)
	return err
}

// IsBoardModerator 判断用户是否为该板版主。
func IsBoardModerator(userID, boardID int64) (bool, error) {
	var n int64
	err := db.Get(&n, `
		SELECT COUNT(*) FROM board_moderator WHERE user_id = $1 AND board_id = $2`,
		userID, boardID,
	)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// ListModeratedBoardIDsByUser 返回用户担任版主的所有板块 id。
func ListModeratedBoardIDsByUser(userID int64) ([]int64, error) {
	var ids []int64
	err := db.Select(&ids, `
		SELECT board_id FROM board_moderator WHERE user_id = $1 ORDER BY board_id ASC`, userID)
	if err != nil {
		return nil, err
	}
	return ids, nil
}
