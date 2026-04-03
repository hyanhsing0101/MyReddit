package logic

import (
	"myreddit/models"
	"myreddit/dao/postgres"
	"time"
	"database/sql"
)

func CreatePost(p *models.ParamCreatePost, userID int64) error {
	post := models.Post{
		Title: p.Title,
		Content: p.Content,
		AuthorID: sql.NullInt64{Int64: userID, Valid: true},
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
	if err := postgres.CreatePost(&post); err != nil {
		return err
	}
	return nil
}