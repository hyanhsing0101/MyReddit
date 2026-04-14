package logic

import (
	"myreddit/dao/postgres"
	"myreddit/models"
)

// ListTags 分页获取标签列表。
func ListTags(p *models.ParamTagList) (*models.TagListData, error) {
	p.Normalize()
	total, err := postgres.CountTags()
	if err != nil {
		return nil, err
	}
	offset := (p.Page - 1) * p.PageSize
	list, err := postgres.ListTags(p.PageSize, offset)
	if err != nil {
		return nil, err
	}
	return &models.TagListData{
		List:     list,
		Total:    total,
		Page:     p.Page,
		PageSize: p.PageSize,
	}, nil
}
