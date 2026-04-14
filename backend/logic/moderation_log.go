package logic

import (
	"myreddit/dao/postgres"
	"myreddit/models"
	"time"
)

func appendModerationLog(boardID, operatorID int64, action models.ModerationAction, targetType models.ModerationTargetType, targetID int64, description string) {
	_ = postgres.CreateModerationLog(
		boardID,
		operatorID,
		string(action),
		string(targetType),
		targetID,
		description,
	)
}

func ListBoardModerationLogs(boardID, operatorID int64, p *models.ParamModerationLogList) (*models.ModerationLogListData, error) {
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
	total, err := postgres.CountBoardModerationLogs(boardID, p.Action, p.TargetType, p.TargetID)
	if err != nil {
		return nil, err
	}
	offset := (p.Page - 1) * p.PageSize
	list, err := postgres.ListBoardModerationLogs(boardID, p.Action, p.TargetType, p.TargetID, p.PageSize, offset)
	if err != nil {
		return nil, err
	}
	return &models.ModerationLogListData{
		List:     list,
		Total:    total,
		Page:     p.Page,
		PageSize: p.PageSize,
	}, nil
}

func GetModerationDashboard(boardID, operatorID int64) (*models.ModerationDashboardData, error) {
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
	now := time.Now()
	since24h := now.Add(-24 * time.Hour)
	since7d := now.Add(-7 * 24 * time.Hour)

	pending, err := postgres.CountBoardPostReports(boardID, string(models.PostReportStatusOpen))
	if err != nil {
		return nil, err
	}
	inReview, err := postgres.CountBoardPostReports(boardID, string(models.PostReportStatusReview))
	if err != nil {
		return nil, err
	}
	resolved7d, err := postgres.CountBoardPostReportsSince(boardID, string(models.PostReportStatusResolved), since7d)
	if err != nil {
		return nil, err
	}
	rejected7d, err := postgres.CountBoardPostReportsSince(boardID, string(models.PostReportStatusRejected), since7d)
	if err != nil {
		return nil, err
	}
	logs24h, err := postgres.CountBoardModerationLogsSince(boardID, since24h)
	if err != nil {
		return nil, err
	}
	reports24h, err := postgres.CountBoardPostReportsSince(boardID, "", since24h)
	if err != nil {
		return nil, err
	}
	return &models.ModerationDashboardData{
		PendingReports:    pending,
		InReviewReports:   inReview,
		ResolvedReports7d: resolved7d,
		RejectedReports7d: rejected7d,
		Logs24h:           logs24h,
		ReportsCreated24h: reports24h,
	}, nil
}
