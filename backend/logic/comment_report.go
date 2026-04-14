package logic

import (
	"errors"
	"fmt"
	"myreddit/dao/postgres"
	"myreddit/models"
	"strings"
	"time"
)

var (
	ErrCannotReportOwnComment       = errors.New("cannot report own comment")
	ErrDuplicateActiveCommentReport = errors.New("duplicate active comment report")
)

func CreateCommentReport(postID, commentID, reporterID int64, p *models.ParamCreatePostReport) error {
	r, err := postReader(&reporterID)
	if err != nil {
		return err
	}
	post, err := postgres.GetPostByID(postID, r)
	if err != nil {
		return err
	}
	comment, err := postgres.GetActiveCommentByPostAndID(postID, commentID)
	if err != nil {
		return err
	}
	if comment.AuthorID.Valid && comment.AuthorID.Int64 == reporterID {
		return ErrCannotReportOwnComment
	}
	dup, err := postgres.HasActiveCommentReportByUser(commentID, reporterID)
	if err != nil {
		return err
	}
	if dup {
		return ErrDuplicateActiveCommentReport
	}
	reason := strings.TrimSpace(p.Reason)
	detail := strings.TrimSpace(p.Detail)
	return postgres.CreateCommentReport(commentID, postID, post.BoardID, reporterID, reason, detail, time.Now())
}

func ListBoardCommentReports(boardID, operatorID int64, p *models.ParamReportList) (*models.CommentReportListData, error) {
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
	pendingOpen, err := postgres.CountBoardCommentReports(boardID, string(models.CommentReportStatusOpen))
	if err != nil {
		return nil, err
	}
	total, err := postgres.CountBoardCommentReports(boardID, p.Status)
	if err != nil {
		return nil, err
	}
	offset := (p.Page - 1) * p.PageSize
	list, err := postgres.ListBoardCommentReports(boardID, p.Status, p.PageSize, offset)
	if err != nil {
		return nil, err
	}
	return &models.CommentReportListData{
		List:        list,
		Total:       total,
		PendingOpen: pendingOpen,
		Page:        p.Page,
		PageSize:    p.PageSize,
	}, nil
}

func UpdateCommentReportStatus(reportID, operatorID int64, p *models.ParamUpdatePostReport) error {
	report, err := postgres.GetCommentReportByID(reportID)
	if err != nil {
		return err
	}
	ok, err := canModerateBoard(operatorID, report.BoardID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrManageReportForbidden
	}
	if err := postgres.UpdateCommentReportStatus(
		reportID,
		p.Status,
		operatorID,
		strings.TrimSpace(p.HandlerNote),
		time.Now(),
	); err != nil {
		return err
	}
	appendModerationLog(
		report.BoardID,
		operatorID,
		models.ModerationActionResolveCommentReport,
		models.ModerationTargetCommentReport,
		reportID,
		fmt.Sprintf("status=%s", p.Status),
	)
	return nil
}

func BatchUpdateCommentReportStatus(boardID, operatorID int64, p *models.ParamBatchUpdateCommentReports) error {
	if _, err := postgres.GetBoardByID(boardID); err != nil {
		return err
	}
	ok, err := canModerateBoard(operatorID, boardID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrManageReportForbidden
	}
	count, err := postgres.CountCommentReportsByIDsInBoard(boardID, p.ReportIDs)
	if err != nil {
		return err
	}
	if count != int64(len(p.ReportIDs)) {
		return postgres.ErrorCommentReportNotExist
	}
	note := strings.TrimSpace(p.HandlerNote)
	if err := postgres.BatchUpdateCommentReportStatus(
		boardID,
		p.ReportIDs,
		p.Status,
		operatorID,
		note,
		time.Now(),
	); err != nil {
		return err
	}
	for _, id := range p.ReportIDs {
		appendModerationLog(
			boardID,
			operatorID,
			models.ModerationActionResolveCommentReport,
			models.ModerationTargetCommentReport,
			id,
			fmt.Sprintf("status=%s,batch=true", p.Status),
		)
	}
	return nil
}
