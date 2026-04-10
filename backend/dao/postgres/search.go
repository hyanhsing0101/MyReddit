package postgres

import (
	"myreddit/models"
)

// SearchPostsByFTS 搜帖子（标题+正文），仅返回当前读者可见的未软删帖子。
func SearchPostsByFTS(q string, limit int, r *PostReader) ([]models.SearchPostItem, error) {
	list := make([]models.SearchPostItem, 0)
	if q == "" {
		return list, nil
	}
	uid := postReaderUID(r)
	adm := postReaderAdmin(r)
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
    ` + postReadableWhere(3, 4) + `
    AND p.search_vector @@ query.tsq
ORDER BY score DESC, p.create_time DESC
LIMIT $2;
`
	err := db.Select(&list, sqlStr, q, limit, uid, adm)
	if err != nil {
		return nil, err
	}

	return list, nil
}

// SearchBoardsByFTS 搜板块；私有板仅成员/版主/站主可见。
func SearchBoardsByFTS(q string, limit int, viewerID *int64, viewerIsSiteAdmin bool) ([]models.SearchBoardItem, error) {
	list := make([]models.SearchBoardItem, 0)
	if q == "" {
		return list, nil
	}
	var uid interface{}
	if viewerID != nil {
		uid = *viewerID
	} else {
		uid = nil
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
    b.visibility,
    b.is_system_sink,
    b.create_time,
    b.update_time,
    ts_rank_cd(b.search_vector, query.tsq) AS score
FROM "board" b
JOIN query ON true
WHERE
    b.is_system_sink = false
    AND b.search_vector @@ query.tsq
    AND ` + sqlBoardListVisible(3, 4) + `
ORDER BY score DESC, b.create_time DESC
LIMIT $2;
`
	err := db.Select(&list, sqlStr, q, limit, uid, viewerIsSiteAdmin)
	if err != nil {
		return nil, err
	}

	return list, nil
}
