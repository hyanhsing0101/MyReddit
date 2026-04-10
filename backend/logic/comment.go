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
	r, err := postReader(&userID)
	if err != nil {
		return err
	}
	row, err := postgres.GetPostByID(postID, r)
	if err != nil {
		return err
	}
	if row.SealedAt.Valid {
		return ErrPostSealed
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

func ListComments(postID int64, param *models.ParamCommentList, viewerID *int64) (*models.CommentListData, error) {
	r, err := postReader(viewerID)
	if err != nil {
		return nil, err
	}
	if _, err := postgres.GetPostByID(postID, r); err != nil {
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
	if viewerID != nil && len(list) > 0 {
		ids := make([]int64, len(list))
		for i := range list {
			ids[i] = list[i].ID
		}
		m, err := postgres.GetCommentVotesForUser(*viewerID, ids)
		if err != nil {
			return nil, err
		}
		for i := range list {
			if v, ok := m[list[i].ID]; ok {
				x := v
				list[i].MyVote = &x
			}
		}
	}
	return &models.CommentListData{
		List:     list,
		Total:    total,
		Page:     param.Page,
		PageSize: param.PageSize,
	}, nil
}

// VoteComment 对评论上/下票或取消；返回最新 score 与 my_vote。
func VoteComment(postID, commentID, userID int64, value int8) (*models.PostVoteResult, error) {
	r, err := postReader(&userID)
	if err != nil {
		return nil, err
	}
	if _, err := postgres.GetPostByID(postID, r); err != nil {
		return nil, err
	}
	full, err := postgres.GetPostByIDIncludingDeleted(postID)
	if err != nil {
		return nil, err
	}
	if full.SealedAt.Valid {
		return nil, ErrPostSealed
	}
	score, myVote, err := postgres.ApplyCommentVote(postID, commentID, userID, value)
	if err != nil {
		return nil, err
	}
	return &models.PostVoteResult{
		Score:  score,
		MyVote: myVote,
	}, nil
}
