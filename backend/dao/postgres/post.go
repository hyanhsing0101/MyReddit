package postgres

import (
	"database/sql"
	"errors"
	"myreddit/models"
)

var ErrorPostNotExist = errors.New("post not exist")

func CreatePost(post *models.Post) error {
	sqlStr := `insert into "post" (board_id, title, content, author_id, create_time, update_time) values ($1, $2, $3, $4, $5, $6)`
	_, err := db.Exec(sqlStr, post.BoardID, post.Title, post.Content, post.AuthorID, post.CreateTime, post.UpdateTime)
	return err
}

func CountPosts(boardID *int64) (int64, error) {
	var n int64
	var err error
	if boardID == nil {
		err = db.Get(&n, `select count(*) from "post"`)
	} else {
		err = db.Get(&n, `select count(*) from "post" where board_id = $1`, *boardID)
	}
	return n, err
}

func ListPosts(boardID *int64, limit, offset int) ([]models.Post, error) {
	var list []models.Post
	var err error
	base := `
		select p.id, p.board_id, p.title, p.content, p.author_id, p.create_time, p.update_time,
		       b.slug as board_slug, b.name as board_name
		from "post" p
		inner join "board" b on b.id = p.board_id`
	if boardID == nil {
		sqlStr := base + ` order by p.create_time desc limit $1 offset $2`
		err = db.Select(&list, sqlStr, limit, offset)
	} else {
		sqlStr := base + ` where p.board_id = $1 order by p.create_time desc limit $2 offset $3`
		err = db.Select(&list, sqlStr, *boardID, limit, offset)
	}
	return list, err
}

func GetPostByID(id int64) (*models.Post, error) {
	var p models.Post
	sqlStr := `
		select p.id, p.board_id, p.title, p.content, p.author_id, p.create_time, p.update_time,
		       b.slug as board_slug, b.name as board_name
		from "post" p
		inner join "board" b on b.id = p.board_id
		where p.id = $1`
	err := db.Get(&p, sqlStr, id)
	if err == sql.ErrNoRows {
		return nil, ErrorPostNotExist
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}
