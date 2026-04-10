package postgres

import (
	"database/sql"
	"myreddit/models"
	"time"
)

// InsertBoardModerator 写入版主行（建板事务内调用）。
func InsertBoardModerator(userID, boardID int64, role string, appointedBy int64, at time.Time) error {
	_, err := db.Exec(`
		INSERT INTO board_moderator (user_id, board_id, role, appointed_by, create_time, update_time)
		VALUES ($1, $2, $3, $4, $5, $5)
		ON CONFLICT (user_id, board_id) DO NOTHING`,
		userID, boardID, role, appointedBy, at,
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

// IsBoardOwner 判断用户是否为该板 owner。
func IsBoardOwner(userID, boardID int64) (bool, error) {
	var n int64
	err := db.Get(&n, `
		SELECT COUNT(*) FROM board_moderator
		WHERE user_id = $1 AND board_id = $2 AND role = $3`,
		userID, boardID, models.BoardModeratorRoleOwner,
	)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func CountBoardOwners(boardID int64) (int64, error) {
	var n int64
	err := db.Get(&n, `
		SELECT COUNT(*) FROM board_moderator WHERE board_id = $1 AND role = $2`,
		boardID, models.BoardModeratorRoleOwner,
	)
	return n, err
}

func GetBoardModerator(boardID, userID int64) (*models.BoardModerator, error) {
	var row models.BoardModerator
	err := db.Get(&row, `
		SELECT bm.user_id, bm.board_id, bm.role, u.username, bm.appointed_by, bm.create_time, bm.update_time
		FROM board_moderator bm
		LEFT JOIN "user" u ON u.user_id = bm.user_id
		WHERE bm.board_id = $1 AND bm.user_id = $2
	`, boardID, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func ListBoardModerators(boardID int64) ([]models.BoardModerator, error) {
	var list []models.BoardModerator
	err := db.Select(&list, `
		SELECT bm.user_id, bm.board_id, bm.role, u.username, bm.appointed_by, bm.create_time, bm.update_time
		FROM board_moderator bm
		LEFT JOIN "user" u ON u.user_id = bm.user_id
		WHERE bm.board_id = $1
		ORDER BY (bm.role = 'owner') DESC, bm.create_time ASC, bm.user_id ASC
	`, boardID)
	return list, err
}

func UpsertBoardModerator(boardID, userID int64, role string, appointedBy int64, at time.Time) error {
	_, err := db.Exec(`
		INSERT INTO board_moderator (user_id, board_id, role, appointed_by, create_time, update_time)
		VALUES ($1, $2, $3, $4, $5, $5)
		ON CONFLICT (user_id, board_id) DO UPDATE
		SET role = EXCLUDED.role,
		    appointed_by = EXCLUDED.appointed_by,
		    update_time = EXCLUDED.update_time
	`, userID, boardID, role, appointedBy, at)
	return err
}

func UpdateBoardModeratorRole(boardID, userID int64, role string, appointedBy int64, at time.Time) (bool, error) {
	res, err := db.Exec(`
		UPDATE board_moderator
		SET role = $3, appointed_by = $4, update_time = $5
		WHERE board_id = $1 AND user_id = $2
	`, boardID, userID, role, appointedBy, at)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func RemoveBoardModerator(boardID, userID int64) (bool, error) {
	res, err := db.Exec(`DELETE FROM board_moderator WHERE board_id = $1 AND user_id = $2`, boardID, userID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
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
