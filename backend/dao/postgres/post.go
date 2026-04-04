package postgres

import (
	"database/sql"
	"errors"
	"myreddit/models"
)

var ErrorPostNotExist = errors.New("post not exist")

func CreatePost(post *models.Post) error {
	sqlStr := `insert into "post" (title, content, author_id, create_time, update_time) values ($1, $2, $3, $4, $5)`
	_, err := db.Exec(sqlStr, post.Title, post.Content, post.AuthorID, post.CreateTime, post.UpdateTime)
	return err
}

func CountPosts() (int64, error) {
	var n int64
	sqlStr := `select count(*) from "post"`
	err := db.Get(&n, sqlStr)
	return n, err
}

func ListPosts(limit, offset int) ([]models.Post, error) {
	var list []models.Post
	sqlStr := `
		select id, title, content, author_id, create_time, update_time
		from "post"
		order by create_time desc
		limit $1 offset $2`
	err := db.Select(&list, sqlStr, limit, offset)
	return list, err
}

func GetPostByID(id int64) (*models.Post, error) {
	var p models.Post
	sqlStr := `
		select id, title, content, author_id, create_time, update_time
		from "post"
		where id = $1`
	err := db.Get(&p, sqlStr, id)
	if err == sql.ErrNoRows {
		return nil, ErrorPostNotExist
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}