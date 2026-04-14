package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"myreddit/models"
	"time"
)

var ErrorCommentReportNotExist = errors.New("comment report not exist")

func CountBoardCommentReports(boardID int64, status string) (int64, error) {
	var n int64
	if status == "" {
		err := db.Get(&n, `SELECT COUNT(*) FROM comment_report WHERE board_id = $1`, boardID)
		return n, err
	}
	err := db.Get(&n, `SELECT COUNT(*) FROM comment_report WHERE board_id = $1 AND status = $2`, boardID, status)
	return n, err
}

func ListBoardCommentReports(boardID int64, status string, limit, offset int) ([]models.CommentReportView, error) {
	base := `
		SELECT
			r.id, r.comment_id, r.post_id, COALESCE(p.title, '') AS post_title,
			LEFT(COALESCE(c.content, ''), 200) AS comment_snippet,
			r.board_id,
			r.reporter_id, COALESCE(u.username, '') AS reporter_username,
			r.reason, r.detail, r.status, r.handler_id,
			COALESCE(h.username, '') AS handler_username,
			r.handler_note, r.create_time, r.update_time
		FROM comment_report r
		LEFT JOIN "post" p ON p.id = r.post_id
		LEFT JOIN "comment" c ON c.id = r.comment_id
		LEFT JOIN "user" u ON u.user_id = r.reporter_id
		LEFT JOIN "user" h ON h.user_id = r.handler_id
		WHERE r.board_id = $1`
	args := []interface{}{boardID}
	if status != "" {
		base += ` AND r.status = $2`
		args = append(args, status, limit, offset)
		base += ` ORDER BY r.create_time DESC LIMIT $3 OFFSET $4`
	} else {
		args = append(args, limit, offset)
		base += ` ORDER BY r.create_time DESC LIMIT $2 OFFSET $3`
	}
	var out []models.CommentReportView
	if err := db.Select(&out, base, args...); err != nil {
		return nil, err
	}
	return out, nil
}

func HasActiveCommentReportByUser(commentID, reporterID int64) (bool, error) {
	var n int64
	err := db.Get(&n, `
		SELECT COUNT(*) FROM comment_report
		WHERE comment_id = $1 AND reporter_id = $2 AND status IN ('open', 'in_review')`,
		commentID, reporterID,
	)
	return n > 0, err
}

func CreateCommentReport(commentID, postID, boardID, reporterID int64, reason, detail string, now time.Time) error {
	_, err := db.Exec(`
		INSERT INTO comment_report (comment_id, post_id, board_id, reporter_id, reason, detail, status, create_time, update_time)
		VALUES ($1, $2, $3, $4, $5, $6, 'open', $7, $7)`,
		commentID, postID, boardID, reporterID, reason, detail, now,
	)
	return err
}

func GetCommentReportByID(id int64) (*models.CommentReport, error) {
	var r models.CommentReport
	err := db.Get(&r, `
		SELECT id, comment_id, post_id, board_id, reporter_id, reason, detail, status, handler_id, handler_note, create_time, update_time
		FROM comment_report
		WHERE id = $1`,
		id,
	)
	if err == sql.ErrNoRows {
		return nil, ErrorCommentReportNotExist
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func UpdateCommentReportStatus(reportID int64, status string, handlerID int64, handlerNote string, now time.Time) error {
	res, err := db.Exec(`
		UPDATE comment_report
		SET status = $2, handler_id = $3, handler_note = $4, update_time = $5
		WHERE id = $1`,
		reportID, status, handlerID, handlerNote, now,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrorCommentReportNotExist
	}
	return nil
}

func CountCommentReportsByIDsInBoard(boardID int64, ids []int64) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	var n int64
	err := db.Get(&n, `
		SELECT COUNT(*) FROM comment_report
		WHERE board_id = $1 AND id = ANY($2::bigint[])`,
		boardID, Int64Array(ids),
	)
	return n, err
}

func BatchUpdateCommentReportStatus(boardID int64, ids []int64, status string, handlerID int64, handlerNote string, now time.Time) error {
	if len(ids) == 0 {
		return nil
	}
	sqlStr := `
		UPDATE comment_report
		SET status = $3, handler_id = $4, handler_note = $5, update_time = $6
		WHERE board_id = $1 AND id = ANY($2::bigint[])`
	res, err := db.Exec(sqlStr, boardID, Int64Array(ids), status, handlerID, handlerNote, now)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != int64(len(ids)) {
		return fmt.Errorf("batch update mismatch: want=%d got=%d", len(ids), n)
	}
	return nil
}
