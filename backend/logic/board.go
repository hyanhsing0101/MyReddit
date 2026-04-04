package logic

import (
	"database/sql"
	"errors"
	"myreddit/dao/postgres"
	"myreddit/models"
	"regexp"
	"strings"
	"time"
)

const boardSlugReserved = "_archived"

var boardSlugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_]{0,62}$`)

var (
	ErrBoardSlugReserved = errors.New("board slug reserved")
	ErrBoardSlugInvalid  = errors.New("board slug invalid")
	ErrBoardNameEmpty    = errors.New("board name empty")
)

func normalizeBoardSlug(slug string) string {
	return strings.ToLower(strings.TrimSpace(slug))
}

func validateBoardSlug(slug string) error {
	if slug == boardSlugReserved {
		return ErrBoardSlugReserved
	}
	if !boardSlugPattern.MatchString(slug) {
		return ErrBoardSlugInvalid
	}
	return nil
}

func ListBoards(p *models.ParamBoardList) (*models.BoardListData, error) {
	p.Normalize()
	include := p.IncludeSystemSink
	total, err := postgres.CountBoards(include)
	if err != nil {
		return nil, err
	}
	offset := (p.Page - 1) * p.PageSize
	rows, err := postgres.ListBoards(include, p.PageSize, offset)
	if err != nil {
		return nil, err
	}
	list := make([]models.BoardView, 0, len(rows))
	for _, row := range rows {
		list = append(list, models.BoardToView(row))
	}
	return &models.BoardListData{
		List:     list,
		Total:    total,
		Page:     p.Page,
		PageSize: p.PageSize,
	}, nil
}

func GetBoardByID(id int64) (*models.BoardView, error) {
	b, err := postgres.GetBoardByID(id)
	if err != nil {
		return nil, err
	}
	v := models.BoardToView(*b)
	return &v, nil
}

func GetBoardBySlug(slug string) (*models.BoardView, error) {
	slug = normalizeBoardSlug(slug)
	if slug == "" {
		return nil, postgres.ErrorBoardNotExist
	}
	b, err := postgres.GetBoardBySlug(slug)
	if err != nil {
		return nil, err
	}
	v := models.BoardToView(*b)
	return &v, nil
}

func CreateBoard(p *models.ParamCreateBoard, userID int64) error {
	slug := normalizeBoardSlug(p.Slug)
	if err := validateBoardSlug(slug); err != nil {
		return err
	}
	name := strings.TrimSpace(p.Name)
	if name == "" {
		return ErrBoardNameEmpty
	}
	desc := strings.TrimSpace(p.Description)
	var descNull sql.NullString
	if desc != "" {
		descNull = sql.NullString{String: desc, Valid: true}
	}
	now := time.Now()
	b := models.Board{
		Slug:         slug,
		Name:         name,
		Description:  descNull,
		CreatedBy:    sql.NullInt64{Int64: userID, Valid: true},
		CreateTime:   now,
		UpdateTime:   now,
	}
	return postgres.CreateBoard(&b)
}
