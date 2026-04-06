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
	ErrDeletePostForbidden     = errors.New("delete post forbidden")
	ErrTagCountExceedsMaxLimit = errors.New("tag count exceeds the maximum limit")
)

const MaxPostTagCount = 5

func CreatePost(p *models.ParamCreatePost, userID int64) error {
	board, err := postgres.GetBoardByID(p.BoardID)
	if err != nil {
		return err
	}
	if board.IsSystemSink {
		return ErrCannotPostToSystemBoard
	}
	tagIDs, err := postgres.ValidateTagIDs(p.TagIDs)
	if err != nil {
		return err
	}
	if len(tagIDs) > MaxPostTagCount {
		return ErrTagCountExceedsMaxLimit
	}
	post := models.Post{
		BoardID:    p.BoardID,
		Title:      p.Title,
		Content:    p.Content,
		AuthorID:   sql.NullInt64{Int64: userID, Valid: true},
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
	postid, err := postgres.CreatePost(&post)
	if err != nil {
		return err
	}
	if err := postgres.ReplacePostTags(postid, tagIDs); err != nil {
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
		v := models.PostToView(row)
		Tags, err := postgres.GetTagsByPostID(row.ID)
		if err != nil {
			return nil, err
		}
		v.Tags = Tags
		list = append(list, v)
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
	Tags, err := postgres.GetTagsByPostID(post.ID)
	if err != nil {
		return nil, err
	}
	v.Tags = Tags
	return &v, nil
}

// DeletePost 软删：作者本人或站点管理员；无主帖仅管理员可删；已软删返回 post not exist
func DeletePost(postID, operatorUserID int64) error {
	post, err := postgres.GetPostByIDIncludingDeleted(postID)
	if err != nil {
		return err
	}
	if post.DeletedAt.Valid {
		return postgres.ErrorPostNotExist
	}
	admin, err := postgres.IsSiteAdmin(operatorUserID)
	if err != nil {
		return err
	}
	if admin {
		return postgres.SoftDeletePost(postID, time.Now())
	}
	if !post.AuthorID.Valid {
		return ErrDeletePostForbidden
	}
	if post.AuthorID.Int64 != operatorUserID {
		return ErrDeletePostForbidden
	}
	return postgres.SoftDeletePost(postID, time.Now())
}
