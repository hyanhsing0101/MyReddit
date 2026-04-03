package postgres

import (
	"myreddit/models"
)

func CreatePost(post *models.Post) error {
	sqlStr := `insert into post (title, content, author_id, create_time, update_time) values ($1, $2, $3, $4, $5)`
	_, err := db.Exec(sqlStr, post.Title, post.Content, post.AuthorID, post.CreateTime, post.UpdateTime)
	return err
}