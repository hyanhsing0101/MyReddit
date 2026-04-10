package logic

import (
	"errors"
	"myreddit/dao/postgres"
	"myreddit/models"
	"time"
)

var (
	ErrManageBoardModeratorsForbidden = errors.New("manage board moderators forbidden")
	ErrBoardModeratorNotExist         = errors.New("board moderator not exist")
	ErrCannotRemoveLastOwner          = errors.New("cannot remove last owner")
)

func ListBoardModerators(boardID int64, viewerID int64) ([]models.BoardModeratorView, error) {
	b, err := postgres.GetBoardByID(boardID)
	if err != nil {
		return nil, err
	}
	ok, err := canReadBoard(&viewerID, b)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, postgres.ErrorBoardNotExist
	}
	rows, err := postgres.ListBoardModerators(boardID)
	if err != nil {
		return nil, err
	}
	out := make([]models.BoardModeratorView, 0, len(rows))
	for _, r := range rows {
		out = append(out, models.BoardModeratorToView(r))
	}
	return out, nil
}

func AddOrUpdateBoardModerator(boardID, operatorID int64, p *models.ParamAddBoardModerator) error {
	if _, err := postgres.GetBoardByID(boardID); err != nil {
		return err
	}
	ok, err := canManageModerators(operatorID, boardID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrManageBoardModeratorsForbidden
	}
	if _, err := postgres.GetUsernameByUserID(p.UserID); err != nil {
		return err
	}
	now := time.Now()
	return postgres.UpsertBoardModerator(boardID, p.UserID, p.Role, operatorID, now)
}

func UpdateBoardModeratorRole(boardID, targetUserID, operatorID int64, p *models.ParamUpdateBoardModeratorRole) error {
	if _, err := postgres.GetBoardByID(boardID); err != nil {
		return err
	}
	ok, err := canManageModerators(operatorID, boardID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrManageBoardModeratorsForbidden
	}
	old, err := postgres.GetBoardModerator(boardID, targetUserID)
	if err != nil {
		return err
	}
	if old == nil {
		return ErrBoardModeratorNotExist
	}
	if old.Role == models.BoardModeratorRoleOwner && p.Role != models.BoardModeratorRoleOwner {
		n, err := postgres.CountBoardOwners(boardID)
		if err != nil {
			return err
		}
		if n <= 1 {
			return ErrCannotRemoveLastOwner
		}
	}
	okUpdated, err := postgres.UpdateBoardModeratorRole(boardID, targetUserID, p.Role, operatorID, time.Now())
	if err != nil {
		return err
	}
	if !okUpdated {
		return ErrBoardModeratorNotExist
	}
	return nil
}

func RemoveBoardModerator(boardID, targetUserID, operatorID int64) error {
	if _, err := postgres.GetBoardByID(boardID); err != nil {
		return err
	}
	ok, err := canManageModerators(operatorID, boardID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrManageBoardModeratorsForbidden
	}
	row, err := postgres.GetBoardModerator(boardID, targetUserID)
	if err != nil {
		return err
	}
	if row == nil {
		return ErrBoardModeratorNotExist
	}
	if row.Role == models.BoardModeratorRoleOwner {
		n, err := postgres.CountBoardOwners(boardID)
		if err != nil {
			return err
		}
		if n <= 1 {
			return ErrCannotRemoveLastOwner
		}
	}
	okDeleted, err := postgres.RemoveBoardModerator(boardID, targetUserID)
	if err != nil {
		return err
	}
	if !okDeleted {
		return ErrBoardModeratorNotExist
	}
	return nil
}
