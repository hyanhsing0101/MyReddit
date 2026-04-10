package logic

import (
	"myreddit/dao/postgres"
	"myreddit/models"
)

// canReadBoard 私有板：成员、版主、站主可读；公开板与系统板对匿名可读（系统板展示策略由列表接口控制）。
func canReadBoard(viewerID *int64, b *models.Board) (bool, error) {
	if b == nil {
		return false, nil
	}
	if b.IsSystemSink || b.Visibility == models.BoardVisibilityPublic || b.Visibility == "" {
		return true, nil
	}
	if viewerID == nil {
		return false, nil
	}
	admin, err := postgres.IsSiteAdmin(*viewerID)
	if err != nil {
		return false, err
	}
	if admin {
		return true, nil
	}
	mod, err := postgres.IsBoardModerator(*viewerID, b.ID)
	if err != nil {
		return false, err
	}
	if mod {
		return true, nil
	}
	return postgres.HasBoardFavorite(*viewerID, b.ID)
}

// canPostToBoard 公开板任意登录用户可发；私有板须成员/版主/站主。
func canPostToBoard(userID int64, b *models.Board) (bool, error) {
	if b == nil || b.IsSystemSink {
		return false, nil
	}
	if b.Visibility == models.BoardVisibilityPublic || b.Visibility == "" {
		return true, nil
	}
	admin, err := postgres.IsSiteAdmin(userID)
	if err != nil {
		return false, err
	}
	if admin {
		return true, nil
	}
	mod, err := postgres.IsBoardModerator(userID, b.ID)
	if err != nil {
		return false, err
	}
	if mod {
		return true, nil
	}
	return postgres.HasBoardFavorite(userID, b.ID)
}

// canModerateBoard 版主或站主可在该板治理帖子。
func canModerateBoard(userID, boardID int64) (bool, error) {
	admin, err := postgres.IsSiteAdmin(userID)
	if err != nil {
		return false, err
	}
	if admin {
		return true, nil
	}
	return postgres.IsBoardModerator(userID, boardID)
}
