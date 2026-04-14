package logic

import (
	"errors"
	"fmt"
	"myreddit/dao/postgres"
	redisDao "myreddit/dao/redis"
	"myreddit/models"
	"strings"
	"time"
)

var (
	ErrPostAppealForbidden      = errors.New("post appeal forbidden")
	ErrCannotAppealUnsealedPost = errors.New("cannot appeal unsealed post")
)

func UpsertMyPostAppeal(postID, userID int64, p *models.ParamCreatePostAppeal) error {
	post, err := postgres.GetPostByIDIncludingDeleted(postID)
	if err != nil {
		return err
	}
	if post.DeletedAt.Valid {
		return postgres.ErrorPostNotExist
	}
	if !post.AuthorID.Valid || post.AuthorID.Int64 != userID {
		return ErrPostAppealForbidden
	}
	if !post.SealedAt.Valid {
		return ErrCannotAppealUnsealedPost
	}
	reason := strings.TrimSpace(p.Reason)
	reqTitle := strings.TrimSpace(p.RequestedTitle)
	reqBody := strings.TrimSpace(p.RequestedBody)
	userReply := strings.TrimSpace(p.UserReply)
	now := time.Now()

	latest, err := postgres.GetLatestPostAppealByPostAndAuthor(postID, userID)
	if err == nil && (latest.Status == models.PostAppealStatusOpen || latest.Status == models.PostAppealStatusReview) {
		return postgres.UpdatePostAppealByID(latest.ID, reason, reqTitle, reqBody, userReply, now)
	}
	if err != nil && !errors.Is(err, postgres.ErrorPostAppealNotExist) {
		return err
	}
	_, err = postgres.CreatePostAppeal(postID, post.BoardID, userID, reason, reqTitle, reqBody, userReply, now)
	return err
}

func GetMyPostAppeal(postID, userID int64) (*models.PostAppealView, error) {
	post, err := postgres.GetPostByIDIncludingDeleted(postID)
	if err != nil {
		return nil, err
	}
	if post.DeletedAt.Valid {
		return nil, postgres.ErrorPostNotExist
	}
	if !post.AuthorID.Valid || post.AuthorID.Int64 != userID {
		return nil, ErrPostAppealForbidden
	}
	row, err := postgres.GetLatestPostAppealByPostAndAuthor(postID, userID)
	if err != nil {
		return nil, err
	}
	return &models.PostAppealView{
		ID:             row.ID,
		PostID:         row.PostID,
		PostTitle:      post.Title,
		BoardID:        row.BoardID,
		AuthorID:       row.AuthorID,
		Reason:         row.Reason,
		RequestedTitle: row.RequestedTitle,
		RequestedBody:  row.RequestedBody,
		UserReply:      row.UserReply,
		Status:         row.Status,
		ModeratorID:    row.ModeratorID,
		ModeratorReply: row.ModeratorReply,
		CreateTime:     row.CreateTime,
		UpdateTime:     row.UpdateTime,
	}, nil
}

func ListBoardPostAppeals(boardID, operatorID int64, p *models.ParamPostAppealList) (*models.PostAppealListData, error) {
	if _, err := postgres.GetBoardByID(boardID); err != nil {
		return nil, err
	}
	ok, err := canModerateBoard(operatorID, boardID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrManageReportForbidden
	}
	p.Normalize()
	total, err := postgres.CountBoardPostAppeals(boardID, p.Status)
	if err != nil {
		return nil, err
	}
	offset := (p.Page - 1) * p.PageSize
	list, err := postgres.ListBoardPostAppeals(boardID, p.Status, p.PageSize, offset)
	if err != nil {
		return nil, err
	}
	return &models.PostAppealListData{
		List:     list,
		Total:    total,
		Page:     p.Page,
		PageSize: p.PageSize,
	}, nil
}

func HandlePostAppeal(appealID, operatorID int64, p *models.ParamHandlePostAppeal) error {
	row, err := postgres.GetPostAppealByID(appealID)
	if err != nil {
		return err
	}
	ok, err := canModerateBoard(operatorID, row.BoardID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrManageReportForbidden
	}
	now := time.Now()
	reply := strings.TrimSpace(p.ModeratorReply)
	if err := postgres.HandlePostAppeal(appealID, p.Status, operatorID, reply, now); err != nil {
		return err
	}
	post, err := postgres.GetPostByIDIncludingDeleted(row.PostID)
	if err != nil {
		return err
	}
	if p.Status == string(models.PostAppealStatusApproved) {
		if p.ApplyUpdate {
			if err := postgres.UpdatePostContent(row.PostID, row.RequestedTitle, row.RequestedBody, now); err != nil {
				return err
			}
		}
		if post.SealedAt.Valid {
			if err := postgres.UnsealPost(row.PostID, now); err != nil {
				return err
			}
		}
		if fresh, err := postgres.GetPostByIDIncludingDeleted(row.PostID); err == nil && !fresh.DeletedAt.Valid {
			_ = redisDao.UpsertHotPost(row.PostID, calcHotScore(fresh.Score, fresh.CreateTime))
		}
	}
	appendModerationLog(
		row.BoardID,
		operatorID,
		models.ModerationActionHandlePostAppeal,
		models.ModerationTargetPost,
		row.PostID,
		fmt.Sprintf("appeal_id=%d,status=%s,apply_update=%t", row.ID, p.Status, p.ApplyUpdate),
	)
	return nil
}
