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
	ErrBoardNameEmpty         = errors.New("board name empty")
	ErrBoardVisibilityInvalid = errors.New("board visibility invalid")
	// ErrCannotFavoriteSystemBoard 系统归档板等不可收藏
	ErrCannotFavoriteSystemBoard = errors.New("cannot favorite system board")
	// ErrCannotFavoritePublicBoard 公开板不提供订阅/收藏
	ErrCannotFavoritePublicBoard = errors.New("cannot favorite public board")
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

func attachFavoriteFlagsToBoards(list []models.BoardView, viewerID *int64) error {
	if viewerID == nil || len(list) == 0 {
		return nil
	}
	ids := make([]int64, len(list))
	for i := range list {
		ids[i] = list[i].ID
	}
	found, err := postgres.ListBoardIDsFavoritedByUser(*viewerID, ids)
	if err != nil {
		return err
	}
	for i := range list {
		if _, ok := found[list[i].ID]; ok {
			t := true
			list[i].IsFavorited = &t
		} else {
			f := false
			list[i].IsFavorited = &f
		}
	}
	return nil
}

func ListBoards(p *models.ParamBoardList, viewerID *int64) (*models.BoardListData, error) {
	p.Normalize()
	include := p.IncludeSystemSink
	admin := false
	var err error
	if viewerID != nil {
		admin, err = postgres.IsSiteAdmin(*viewerID)
		if err != nil {
			return nil, err
		}
	}
	total, err := postgres.CountBoards(include, viewerID, admin)
	if err != nil {
		return nil, err
	}
	offset := (p.Page - 1) * p.PageSize
	rows, err := postgres.ListBoards(include, p.PageSize, offset, viewerID, admin)
	if err != nil {
		return nil, err
	}
	list := make([]models.BoardView, 0, len(rows))
	for _, row := range rows {
		list = append(list, models.BoardToView(row))
	}
	if err := attachFavoriteFlagsToBoards(list, viewerID); err != nil {
		return nil, err
	}
	return &models.BoardListData{
		List:     list,
		Total:    total,
		Page:     p.Page,
		PageSize: p.PageSize,
	}, nil
}

func GetBoardByID(id int64, viewerID *int64) (*models.BoardView, error) {
	b, err := postgres.GetBoardByID(id)
	if err != nil {
		return nil, err
	}
	ok, err := canReadBoard(viewerID, b)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, postgres.ErrorBoardNotExist
	}
	v := models.BoardToView(*b)
	if viewerID != nil {
		m, err := postgres.ListBoardIDsFavoritedByUser(*viewerID, []int64{id})
		if err != nil {
			return nil, err
		}
		if _, ok := m[id]; ok {
			t := true
			v.IsFavorited = &t
		} else {
			f := false
			v.IsFavorited = &f
		}
	}
	return &v, nil
}

func GetBoardBySlug(slug string, viewerID *int64) (*models.BoardView, error) {
	slug = normalizeBoardSlug(slug)
	if slug == "" {
		return nil, postgres.ErrorBoardNotExist
	}
	b, err := postgres.GetBoardBySlug(slug)
	if err != nil {
		return nil, err
	}
	return GetBoardByID(b.ID, viewerID)
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
	vis := strings.TrimSpace(p.Visibility)
	if vis == "" {
		vis = models.BoardVisibilityPublic
	}
	if vis != models.BoardVisibilityPublic && vis != models.BoardVisibilityPrivate {
		return ErrBoardVisibilityInvalid
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
		Visibility:   vis,
		CreateTime:   now,
		UpdateTime:   now,
	}
	return postgres.CreateBoardWithModerator(&b, userID)
}

func AddBoardFavorite(userID, boardID int64) error {
	b, err := postgres.GetBoardByID(boardID)
	if err != nil {
		return err
	}
	if b.IsSystemSink {
		return ErrCannotFavoriteSystemBoard
	}
	if b.Visibility == models.BoardVisibilityPublic || b.Visibility == "" {
		return ErrCannotFavoritePublicBoard
	}
	return postgres.AddBoardFavorite(userID, boardID)
}

func RemoveBoardFavorite(userID, boardID int64) error {
	if _, err := postgres.GetBoardByID(boardID); err != nil {
		return err
	}
	return postgres.RemoveBoardFavorite(userID, boardID)
}

func ListMyFavoriteBoards(userID int64, p *models.ParamFavoriteBoardList) (*models.BoardFavoriteListData, error) {
	p.Normalize()
	total, err := postgres.CountBoardFavoritesByUser(userID)
	if err != nil {
		return nil, err
	}
	offset := (p.Page - 1) * p.PageSize
	boards, favTimes, err := postgres.ListBoardFavoritesByUser(userID, p.PageSize, offset)
	if err != nil {
		return nil, err
	}
	list := make([]models.BoardFavoriteView, 0, len(boards))
	for i := range boards {
		bv := models.BoardToView(boards[i])
		t := true
		bv.IsFavorited = &t
		list = append(list, models.BoardFavoriteView{
			BoardView:   bv,
			FavoritedAt: favTimes[i],
		})
	}
	return &models.BoardFavoriteListData{
		List:     list,
		Total:    total,
		Page:     p.Page,
		PageSize: p.PageSize,
	}, nil
}
