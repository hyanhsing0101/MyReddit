package logic

import (
	"database/sql"
	"errors"
	"myreddit/dao/postgres"
	"myreddit/models"
	"time"
)

var (
	ErrParentCommentMismatch = errors.New("parent comment not in this post")
	ErrInvalidCommentParent  = errors.New("invalid parent_id")
)

func CreateComment(postID int64, userID int64, p *models.ParamCreateComment) error {
	if _, err := postgres.GetPostByID(postID); err != nil {
		return err
	}
	var parentID sql.NullInt64
	if p.ParentID != nil {
		if *p.ParentID < 1 {
			return ErrInvalidCommentParent
		}
		parent, err := postgres.GetActiveCommentByID(*p.ParentID)
		if err != nil {
			return err
		}
		if parent.PostID != postID {
			return ErrParentCommentMismatch
		}
		parentID = sql.NullInt64{Int64: *p.ParentID, Valid: true}
	}
	now := time.Now()
	c := models.Comment{
		PostID:     postID,
		AuthorID:   sql.NullInt64{Int64: userID, Valid: true},
		ParentID:   parentID,
		Content:    p.Content,
		CreateTime: now,
		UpdateTime: now,
	}
	return postgres.CreateComment(&c)
}

func ListComments(postID int64, param *models.ParamCommentList) (*models.CommentListData, error) {
	if _, err := postgres.GetPostByID(postID); err != nil {
		return nil, err
	}
	param.Normalize()
	total, err := postgres.CountCommentsByPostID(postID)
	if err != nil {
		return nil, err
	}
	offset := (param.Page - 1) * param.PageSize
	rows, err := postgres.ListCommentsByPostID(postID, param.PageSize, offset)
	if err != nil {
		return nil, err
	}
	list := make([]models.CommentView, 0, len(rows))
	for _, row := range rows {
		list = append(list, models.CommentToView(row))
	}
	return &models.CommentListData{
		List:     list,
		Total:    total,
		Page:     param.Page,
		PageSize: param.PageSize,
	}, nil
}
