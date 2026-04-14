package postgres

import (
	"myreddit/models"
	"strconv"
	"time"
)

func CreateModerationLog(boardID, operatorID int64, action, targetType string, targetID int64, description string) error {
	_, err := db.Exec(`
		INSERT INTO moderation_log (board_id, operator_id, action, target_type, target_id, description, create_time)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())`,
		boardID, operatorID, action, targetType, targetID, description,
	)
	return err
}

func CountBoardModerationLogs(boardID int64, action string, targetType string, targetID *int64) (int64, error) {
	var n int64
	sqlStr := `SELECT COUNT(*) FROM moderation_log WHERE board_id = $1`
	args := []interface{}{boardID}
	idx := 2
	if action != "" {
		sqlStr += ` AND action = $` + strconv.Itoa(idx)
		args = append(args, action)
		idx++
	}
	if targetType != "" {
		sqlStr += ` AND target_type = $` + strconv.Itoa(idx)
		args = append(args, targetType)
		idx++
	}
	if targetID != nil {
		sqlStr += ` AND target_id = $` + strconv.Itoa(idx)
		args = append(args, *targetID)
	}
	err := db.Get(&n, sqlStr, args...)
	return n, err
}

func ListBoardModerationLogs(boardID int64, action string, targetType string, targetID *int64, limit, offset int) ([]models.ModerationLogView, error) {
	base := `
		SELECT
			l.id, l.board_id, l.operator_id, COALESCE(u.username, '') AS operator_username,
			l.action, l.target_type, l.target_id, l.description, l.create_time
		FROM moderation_log l
		LEFT JOIN "user" u ON u.user_id = l.operator_id
		WHERE l.board_id = $1`
	args := []interface{}{boardID}
	idx := 2
	if action != "" {
		base += ` AND l.action = $` + strconv.Itoa(idx)
		args = append(args, action)
		idx++
	}
	if targetType != "" {
		base += ` AND l.target_type = $` + strconv.Itoa(idx)
		args = append(args, targetType)
		idx++
	}
	if targetID != nil {
		base += ` AND l.target_id = $` + strconv.Itoa(idx)
		args = append(args, *targetID)
		idx++
	}
	base += ` ORDER BY l.create_time DESC LIMIT $` + strconv.Itoa(idx) + ` OFFSET $` + strconv.Itoa(idx+1)
	args = append(args, limit, offset)
	var out []models.ModerationLogView
	if err := db.Select(&out, base, args...); err != nil {
		return nil, err
	}
	return out, nil
}

func CountBoardModerationLogsSince(boardID int64, since time.Time) (int64, error) {
	var n int64
	err := db.Get(&n, `
		SELECT COUNT(*) FROM moderation_log
		WHERE board_id = $1 AND create_time >= $2`,
		boardID, since,
	)
	return n, err
}
