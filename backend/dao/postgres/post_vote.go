package postgres

import (
	"database/sql"
	"errors"
	"time"
)

var ErrorInvalidVoteValue = errors.New("invalid vote value")

// GetPostVotesForUser 批量查询当前用户对若干帖子的投票；未投的 post_id 不出现在 map 中。
func GetPostVotesForUser(userID int64, postIDs []int64) (map[int64]int8, error) {
	out := make(map[int64]int8)
	if len(postIDs) == 0 {
		return out, nil
	}
	type row struct {
		PostID int64 `db:"post_id"`
		Value  int8  `db:"value"`
	}
	var rows []row
	err := db.Select(&rows,
		`SELECT post_id, value FROM post_vote WHERE user_id = $1 AND post_id = ANY($2)`,
		userID, Int64Array(postIDs),
	)
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		out[r.PostID] = r.Value
	}
	return out, nil
}

// ApplyPostVote 在同一事务内更新 post_vote 与 post.score；newVal 须为 1、-1 或 0（取消）。
func ApplyPostVote(postID, userID int64, newVal int8) (score int64, myVote *int8, err error) {
	if newVal != -1 && newVal != 0 && newVal != 1 {
		return 0, nil, ErrorInvalidVoteValue
	}

	tx, err := db.Beginx()
	if err != nil {
		return 0, nil, err
	}

	var locked int64
	err = tx.Get(&locked, `
		SELECT id FROM "post" WHERE id = $1 AND deleted_at IS NULL FOR UPDATE
	`, postID)
	if err == sql.ErrNoRows {
		_ = tx.Rollback()
		return 0, nil, ErrorPostNotExist
	}
	if err != nil {
		_ = tx.Rollback()
		return 0, nil, err
	}

	var oldV sql.NullInt64
	err = tx.Get(&oldV, `
		SELECT value FROM post_vote WHERE post_id = $1 AND user_id = $2
	`, postID, userID)
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
			_, err = tx.Exec(`DELETE FROM post_vote WHERE post_id = $1 AND user_id = $2`, postID, userID)
			if err != nil {
				_ = tx.Rollback()
				return 0, nil, err
			}
			delta = int64(-old)
		}
	case 1, -1:
		if old == 0 {
			_, err = tx.Exec(`
				INSERT INTO post_vote (post_id, user_id, value, create_time, update_time)
				VALUES ($1, $2, $3, $4, $4)
			`, postID, userID, newVal, now)
			if err != nil {
				_ = tx.Rollback()
				return 0, nil, err
			}
			delta = int64(newVal)
		} else if old != newVal {
			_, err = tx.Exec(`
				UPDATE post_vote SET value = $1, update_time = $2 WHERE post_id = $3 AND user_id = $4
			`, newVal, now, postID, userID)
			if err != nil {
				_ = tx.Rollback()
				return 0, nil, err
			}
			delta = int64(newVal - old)
		}
	}

	if delta != 0 {
		_, err = tx.Exec(`
			UPDATE "post" SET score = score + $1, update_time = $2 WHERE id = $3
		`, delta, now, postID)
		if err != nil {
			_ = tx.Rollback()
			return 0, nil, err
		}
	}

	err = tx.Get(&score, `SELECT score FROM "post" WHERE id = $1`, postID)
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
