package logic

import (
	"database/sql"
	"errors"
	"myreddit/dao/postgres"
	"myreddit/models"
	"time"
)

var (
	ErrCannotPostToSystemBoard = errors.New("cannot post to system board")
	ErrInvalidBoardID          = errors.New("invalid board id")
)

func CreatePost(p *models.ParamCreatePost, userID int64) error {
	board, err := postgres.GetBoardByID(p.BoardID)
	if err != nil {
		return err
	}
	if board.IsSystemSink {
		return ErrCannotPostToSystemBoard
	}
	post := models.Post{
		BoardID:    p.BoardID,
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
	var boardFilter *int64
	if p.BoardID != nil {
		if *p.BoardID < 1 {
			return nil, ErrInvalidBoardID
		}
		if _, err := postgres.GetBoardByID(*p.BoardID); err != nil {
			return nil, err
		}
		boardFilter = p.BoardID
	}
	total, err := postgres.CountPosts(boardFilter)
	if err != nil {
		return nil, err
	}
	offset := (p.Page - 1) * p.PageSize
	posts, err := postgres.ListPosts(boardFilter, p.PageSize, offset)
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
