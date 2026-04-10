package logic

import (
	"myreddit/dao/postgres"
	"myreddit/models"
	"strings"
)

// Search 全站搜索；viewerID 用于过滤私有板与不可见帖子。
func Search(p *models.ParamSearch, viewerID *int64) (*models.SearchData, error) {
	p.Normalize()
	q := strings.TrimSpace(p.Q)
	data := &models.SearchData{
		Query:  q,
		Scope:  p.Scope,
		Posts:  []models.SearchPostItem{},
		Boards: []models.SearchBoardItem{},
	}
	if q == "" {
		return data, nil
	}
	r, err := postReader(viewerID)
	if err != nil {
		return nil, err
	}
	admin := false
	if viewerID != nil {
		admin, err = postgres.IsSiteAdmin(*viewerID)
		if err != nil {
			return nil, err
		}
	}
	switch p.Scope {
	case "posts":
		posts, err := postgres.SearchPostsByFTS(q, p.PostLimit, r)
		if err != nil {
			return nil, err
		}
		data.Posts = posts
	case "boards":
		boards, err := postgres.SearchBoardsByFTS(q, p.BoardLimit, viewerID, admin)
		if err != nil {
			return nil, err
		}
		data.Boards = boards
	default:
		posts, err := postgres.SearchPostsByFTS(q, p.PostLimit, r)
		if err != nil {
			return nil, err
		}
		data.Posts = posts
		boards, err := postgres.SearchBoardsByFTS(q, p.BoardLimit, viewerID, admin)
		if err != nil {
			return nil, err
		}
		data.Boards = boards
	}
	return data, nil
}