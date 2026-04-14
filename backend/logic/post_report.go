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
	ErrCannotReportOwnPost       = errors.New("cannot report own post")
	ErrDuplicateActivePostReport = errors.New("duplicate active post report")
	ErrManageReportForbidden     = errors.New("manage report forbidden")
)

func CreatePostReport(postID, reporterID int64, p *models.ParamCreatePostReport) error {
	r, err := postReader(&reporterID)
	if err != nil {
		return err
	}
	post, err := postgres.GetPostByID(postID, r)
	if err != nil {
		return err
	}
	if post.AuthorID.Valid && post.AuthorID.Int64 == reporterID {
		return ErrCannotReportOwnPost
	}
	dup, err := postgres.HasActivePostReportByUser(postID, reporterID)
	if err != nil {
		return err
	}
	if dup {
		return ErrDuplicateActivePostReport
	}
	reason := strings.TrimSpace(p.Reason)
	detail := strings.TrimSpace(p.Detail)
	return postgres.CreatePostReport(postID, post.BoardID, reporterID, reason, detail, time.Now())
}

func ListBoardPostReports(boardID, operatorID int64, p *models.ParamReportList) (*models.PostReportListData, error) {
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
	pendingOpen, err := postgres.CountBoardPostReports(boardID, string(models.PostReportStatusOpen))
	if err != nil {
		return nil, err
	}
	total, err := postgres.CountBoardPostReports(boardID, p.Status)
	if err != nil {
		return nil, err
	}
	offset := (p.Page - 1) * p.PageSize
	list, err := postgres.ListBoardPostReports(boardID, p.Status, p.PageSize, offset)
	if err != nil {
		return nil, err
	}
	return &models.PostReportListData{
		List:        list,
		Total:       total,
		PendingOpen: pendingOpen,
		Page:        p.Page,
		PageSize:    p.PageSize,
	}, nil
}

func UpdatePostReportStatus(reportID, operatorID int64, p *models.ParamUpdatePostReport) error {
	report, err := postgres.GetPostReportByID(reportID)
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
	if err := postgres.UpdatePostReportStatus(
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
		models.ModerationActionResolvePostReport,
		models.ModerationTargetPostReport,
		reportID,
		fmt.Sprintf("status=%s", p.Status),
	)
	return nil
}

func BatchUpdatePostReportStatus(boardID, operatorID int64, p *models.ParamBatchUpdatePostReports) error {
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
	count, err := postgres.CountPostReportsByIDsInBoard(boardID, p.ReportIDs)
	if err != nil {
		return err
	}
	if count != int64(len(p.ReportIDs)) {
		return postgres.ErrorPostReportNotExist
	}
	note := strings.TrimSpace(p.HandlerNote)
	if err := postgres.BatchUpdatePostReportStatus(
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
			models.ModerationActionResolvePostReport,
			models.ModerationTargetPostReport,
			id,
			fmt.Sprintf("status=%s,batch=true", p.Status),
		)
	}
	return nil
}
