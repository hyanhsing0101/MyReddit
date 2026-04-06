package logic

import (
	"myreddit/models"
	"strings"
	"myreddit/dao/postgres"
)

func Search(p *models.ParamSearch) (*models.SearchData, error) {
	p.Normalize()
	q := strings.TrimSpace(p.Q)
	data := &models.SearchData{
		Query: q,
		Scope: p.Scope,
		Posts: []models.SearchPostItem{},
		Boards: []models.SearchBoardItem{},
	}
	if q == "" {
		return data, nil
	}
	switch p.Scope {
	case "posts":
		posts, err := postgres.SearchPostsByFTS(q, p.PostLimit)
		if err != nil {
			return nil, err
		}
		data.Posts = posts
	case "boards":
		boards, err := postgres.SearchBoardsByFTS(q, p.BoardLimit)
		if err != nil {
			return nil, err
		}
		data.Boards = boards
	default:
		posts, err := postgres.SearchPostsByFTS(q, p.PostLimit)
		if err != nil {
			return nil, err
		}
		data.Posts = posts
		boards, err := postgres.SearchBoardsByFTS(q, p.BoardLimit)
		if err != nil {
			return nil, err
		}
		data.Boards = boards
	}
	return data, nil
}