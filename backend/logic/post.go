package logic

import (
	"database/sql"
	"myreddit/dao/postgres"
	"myreddit/models"
	"time"
)

func CreatePost(p *models.ParamCreatePost, userID int64) error {
	post := models.Post{
		Title:      p.Title,
		Content:    p.Content,
		AuthorID:   sql.NullInt64{Int64: userID, Valid: true},
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
	if err := postgres.CreatePost(&post); err != nil {
		return err
	}
	return nil
}

func ListPost(p *models.ParamPostList) (*models.PostListData, error) {
	p.Normalize()
	total, err := postgres.CountPosts()
	if err != nil {
		return nil, err
	}
	offset := (p.Page - 1) * p.PageSize
	posts, err := postgres.ListPosts(p.PageSize, offset)
	if err != nil {
		return nil, err
	}
	list := make([]models.PostView, 0, len(posts))
	for _, row := range posts {
		list = append(list, models.PostToView(row))
	}
	return &models.PostListData{
		List:     list,
		Total:    total,
		Page:     p.Page,
		PageSize: p.PageSize,
	}, nil
}

func GetPost(id int64) (*models.PostView, error) {
	post, err := postgres.GetPostByID(id)
	if err != nil {
		return nil, err
	}
	v := models.PostToView(*post)
	return &v, nil
}