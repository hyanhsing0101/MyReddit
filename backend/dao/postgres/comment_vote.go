package postgres

import (
	"database/sql"
	"time"
)

// GetCommentVotesForUser 当前用户对若干评论的投票；未投的 id 不在 map 中。
func GetCommentVotesForUser(userID int64, commentIDs []int64) (map[int64]int8, error) {
	out := make(map[int64]int8)
	if len(commentIDs) == 0 {
		return out, nil
	}
	type row struct {
		CommentID int64 `db:"comment_id"`
		Value     int8  `db:"value"`
	}
	var rows []row
	err := db.Select(&rows,
		`SELECT comment_id, value FROM comment_vote WHERE user_id = $1 AND comment_id = ANY($2)`,
		userID, Int64Array(commentIDs),
	)
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		out[r.CommentID] = r.Value
	}
	return out, nil
}

// ApplyCommentVote 更新 comment_vote 与 comment.score；评论须属于 postID 且未软删。
func ApplyCommentVote(postID, commentID, userID int64, newVal int8) (score int64, myVote *int8, err error) {
	if newVal != -1 && newVal != 0 && newVal != 1 {
		return 0, nil, ErrorInvalidVoteValue
	}

	tx, err := db.Beginx()
	if err != nil {
		return 0, nil, err
	}

	var locked int64
	err = tx.Get(&locked, `
		SELECT id FROM "comment" WHERE id = $1 AND post_id = $2 AND deleted_at IS NULL FOR UPDATE
	`, commentID, postID)
	if err == sql.ErrNoRows {
		_ = tx.Rollback()
		return 0, nil, ErrorCommentNotExist
	}
	if err != nil {
		_ = tx.Rollback()
		return 0, nil, err
	}

	var oldV sql.NullInt64
	err = tx.Get(&oldV, `
		SELECT value FROM comment_vote WHERE comment_id = $1 AND user_id = $2
	`, commentID, userID)
	if err != nil && err != sql.ErrNoRows {
		_ = tx.Rollback()
		return 0, nil, err
	}
	old := int8(0)
	if oldV.Valid {
		old = int8(oldV.Int64)
	}

	now := time.Now()
	delta := int64(0)

	switch newVal {
	case 0:
		if old != 0 {
			_, err = tx.Exec(`DELETE FROM comment_vote WHERE comment_id = $1 AND user_id = $2`, commentID, userID)
			if err != nil {
				_ = tx.Rollback()
				return 0, nil, err
			}
			delta = int64(-old)
		}
	case 1, -1:
		if old == 0 {
			_, err = tx.Exec(`
				INSERT INTO comment_vote (comment_id, user_id, value, create_time, update_time)
				VALUES ($1, $2, $3, $4, $4)
			`, commentID, userID, newVal, now)
			if err != nil {
				_ = tx.Rollback()
				return 0, nil, err
			}
			delta = int64(newVal)
		} else if old != newVal {
			_, err = tx.Exec(`
				UPDATE comment_vote SET value = $1, update_time = $2 WHERE comment_id = $3 AND user_id = $4
			`, newVal, now, commentID, userID)
			if err != nil {
				_ = tx.Rollback()
				return 0, nil, err
			}
			delta = int64(newVal - old)
		}
	}

	if delta != 0 {
		_, err = tx.Exec(`
			UPDATE "comment" SET score = score + $1, update_time = $2 WHERE id = $3
		`, delta, now, commentID)
		if err != nil {
			_ = tx.Rollback()
			return 0, nil, err
		}
	}

	err = tx.Get(&score, `SELECT score FROM "comment" WHERE id = $1`, commentID)
	if err != nil {
		_ = tx.Rollback()
		return 0, nil, err
	}

	if newVal == 0 {
		myVote = nil
	} else {
		v := newVal
		myVote = &v
	}

	if err = tx.Commit(); err != nil {
		return 0, nil, err
	}
	return score, myVote, nil
}
