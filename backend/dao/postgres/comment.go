package postgres

import (
	"database/sql"
	"errors"
	"myreddit/models"
)

var ErrorCommentNotExist = errors.New("comment not exist")

func CreateComment(c *models.Comment) error {
	const q = `
		insert into "comment" (post_id, author_id, parent_id, content, create_time, update_time)
		values ($1, $2, $3, $4, $5, $6)`
	_, err := db.Exec(q, c.PostID, c.AuthorID, c.ParentID, c.Content, c.CreateTime, c.UpdateTime)
	return err
}

func CountCommentsByPostID(postID int64) (int64, error) {
	var n int64
	err := db.Get(&n,
		`select count(*) from "comment" where post_id = $1 and deleted_at is null`,
		postID,
	)
	return n, err
}

func ListCommentsByPostID(postID int64, limit, offset int) ([]models.Comment, error) {
	var list []models.Comment
	const q = `
		select c.id, c.post_id, c.author_id, c.parent_id, c.content, c.deleted_at, c.score, c.create_time, c.update_time,
		       u.username as author_username
		from "comment" c
		left join "user" u on u.user_id = c.author_id
		where c.post_id = $1 and c.deleted_at is null
		order by c.create_time asc
		limit $2 offset $3`
	err := db.Select(&list, q, postID, limit, offset)
	return list, err
}

// GetActiveCommentByID 未软删的评论，用于校验 parent
// GetActiveCommentByPostAndID 校验评论属于指定帖子且未软删。
func GetActiveCommentByPostAndID(postID, commentID int64) (*models.Comment, error) {
	var c models.Comment
	const q = `
		select c.id, c.post_id, c.author_id, c.parent_id, c.content, c.deleted_at, c.score, c.create_time, c.update_time,
		       u.username as author_username
		from "comment" c
		left join "user" u on u.user_id = c.author_id
		where c.id = $1 and c.post_id = $2 and c.deleted_at is null`
	err := db.Get(&c, q, commentID, postID)
	if err == sql.ErrNoRows {
		return nil, ErrorCommentNotExist
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func GetActiveCommentByID(id int64) (*models.Comment, error) {
	var c models.Comment
	const q = `
		select c.id, c.post_id, c.author_id, c.parent_id, c.content, c.deleted_at, c.score, c.create_time, c.update_time,
		       u.username as author_username
		from "comment" c
		left join "user" u on u.user_id = c.author_id
		where c.id = $1 and c.deleted_at is null`
	err := db.Get(&c, q, id)
	if err == sql.ErrNoRows {
		return nil, ErrorCommentNotExist
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}
