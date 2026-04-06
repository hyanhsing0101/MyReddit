package postgres

import (
	"myreddit/models"
)

// SearchPostsByFTS: 搜帖子（标题+正文），只返回未软删帖子
func SearchPostsByFTS(q string, limit int) ([]models.SearchPostItem, error) {
	list := make([]models.SearchPostItem, 0)
	if q == "" {
		return list, nil
	}
	sqlStr := `
WITH query AS (
    SELECT websearch_to_tsquery('simple', $1) AS tsq
)
SELECT
    p.id,
    p.board_id,
    b.slug  AS board_slug,
    b.name  AS board_name,
    p.title,
    p.content,
    p.author_id,
    p.create_time,
    p.update_time,
    ts_rank_cd(p.search_vector, query.tsq) AS score
FROM "post" p
JOIN "board" b ON b.id = p.board_id
JOIN query ON true
WHERE
    p.deleted_at IS NULL
    AND p.search_vector @@ query.tsq
ORDER BY score DESC, p.create_time DESC
LIMIT $2;
`
	err := db.Select(&list, sqlStr, q, limit)
	if err != nil {
		return nil, err
	}

	return list, nil
}

// SearchBoardsByFTS: 搜板块（name/slug/description），默认隐藏系统归档板
func SearchBoardsByFTS(q string, limit int) ([]models.SearchBoardItem, error) {
	list := make([]models.SearchBoardItem, 0)
	if q == "" {
		return list, nil
	}

	sqlStr := `
WITH query AS (
    SELECT websearch_to_tsquery('simple', $1) AS tsq
)
SELECT
    b.id,
    b.slug,
    b.name,
    COALESCE(b.description, '') AS description,
    b.is_system_sink,
    b.create_time,
    b.update_time,
    ts_rank_cd(b.search_vector, query.tsq) AS score
FROM "board" b
JOIN query ON true
WHERE
    b.is_system_sink = false
    AND b.search_vector @@ query.tsq
ORDER BY score DESC, b.create_time DESC
LIMIT $2;
`
	err := db.Select(&list, sqlStr, q, limit)
	if err != nil {
		return nil, err
	}
	return list, nil
}
