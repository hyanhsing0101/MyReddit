package postgres

import (
	"database/sql"
	"errors"
	"myreddit/models"
	"time"
)

var ErrorPostAppealNotExist = errors.New("post appeal not exist")

func GetLatestPostAppealByPostAndAuthor(postID, authorID int64) (*models.PostAppeal, error) {
	row := new(models.PostAppeal)
	err := db.Get(row, `
		SELECT id, post_id, board_id, author_id, reason, requested_title, requested_content, user_reply,
		       status, moderator_id, moderator_reply, create_time, update_time
		FROM post_appeal
		WHERE post_id = $1 AND author_id = $2
		ORDER BY create_time DESC
		LIMIT 1`,
		postID, authorID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrorPostAppealNotExist
	}
	return row, err
}

func CreatePostAppeal(postID, boardID, authorID int64, reason, reqTitle, reqBody, userReply string, now time.Time) (int64, error) {
	var id int64
	err := db.Get(&id, `
		INSERT INTO post_appeal
		    (post_id, board_id, author_id, reason, requested_title, requested_content, user_reply, status, create_time, update_time)
		VALUES ($1,$2,$3,$4,$5,$6,$7,'open',$8,$8)
		RETURNING id`,
		postID, boardID, authorID, reason, reqTitle, reqBody, userReply, now,
	)
	return id, err
}

func UpdatePostAppealByID(appealID int64, reason, reqTitle, reqBody, userReply string, now time.Time) error {
	res, err := db.Exec(`
		UPDATE post_appeal
		SET reason=$2, requested_title=$3, requested_content=$4, user_reply=$5, update_time=$6
		WHERE id=$1`,
		appealID, reason, reqTitle, reqBody, userReply, now,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrorPostAppealNotExist
	}
	return nil
}

func GetPostAppealByID(id int64) (*models.PostAppeal, error) {
	row := new(models.PostAppeal)
	err := db.Get(row, `
		SELECT id, post_id, board_id, author_id, reason, requested_title, requested_content, user_reply,
		       status, moderator_id, moderator_reply, create_time, update_time
		FROM post_appeal
		WHERE id = $1`,
		id,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrorPostAppealNotExist
	}
	return row, err
}

func CountBoardPostAppeals(boardID int64, status string) (int64, error) {
	var n int64
	if status == "" {
		err := db.Get(&n, `SELECT COUNT(*) FROM post_appeal WHERE board_id = $1`, boardID)
		return n, err
	}
	err := db.Get(&n, `SELECT COUNT(*) FROM post_appeal WHERE board_id = $1 AND status = $2`, boardID, status)
	return n, err
}

func ListBoardPostAppeals(boardID int64, status string, limit, offset int) ([]models.PostAppealView, error) {
	list := make([]models.PostAppealView, 0, limit)
	base := `
		SELECT pa.id, pa.post_id, p.title AS post_title, pa.board_id,
		       pa.author_id, COALESCE(ur.username, '') AS author_username,
		       pa.reason, pa.requested_title, pa.requested_content, pa.user_reply,
		       pa.status, pa.moderator_id, COALESCE(um.username, '') AS moderator_username,
		       pa.moderator_reply, pa.create_time, pa.update_time
		FROM post_appeal pa
		LEFT JOIN "post" p ON p.id = pa.post_id
		LEFT JOIN "user" ur ON ur.user_id = pa.author_id
		LEFT JOIN "user" um ON um.user_id = pa.moderator_id
		WHERE pa.board_id = $1`
	if status == "" {
		err := db.Select(&list, base+` ORDER BY pa.create_time DESC LIMIT $2 OFFSET $3`, boardID, limit, offset)
		return list, err
	}
	err := db.Select(&list, base+` AND pa.status = $2 ORDER BY pa.create_time DESC LIMIT $3 OFFSET $4`, boardID, status, limit, offset)
	return list, err
}

func HandlePostAppeal(appealID int64, status string, moderatorID int64, moderatorReply string, now time.Time) error {
	res, err := db.Exec(`
		UPDATE post_appeal
		SET status=$2, moderator_id=$3, moderator_reply=$4, update_time=$5
		WHERE id=$1`,
		appealID, status, moderatorID, moderatorReply, now,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrorPostAppealNotExist
	}
	return nil
}
