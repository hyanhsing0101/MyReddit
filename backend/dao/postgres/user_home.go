package postgres

import "myreddit/models"

// CountPostsByAuthorIDForViewer 统计某用户在「当前浏览者」视角下可见的未软删帖子数。
func CountPostsByAuthorIDForViewer(profileUserID int64, r *PostReader) (int64, error) {
	var n int64
	err := db.Get(&n, `
		SELECT COUNT(*)
		FROM "post" p
		INNER JOIN "board" b ON b.id = p.board_id
		WHERE p.author_id = $1 AND `+postReadableWhere(2, 3),
		profileUserID, postReaderUID(r), postReaderAdmin(r),
	)
	return n, err
}

// ListPostsByAuthorIDForViewer 按创建时间倒序分页；仅含浏览者可读的帖子。
func ListPostsByAuthorIDForViewer(profileUserID int64, r *PostReader, limit, offset int) ([]models.Post, error) {
	var list []models.Post
	sqlStr := `
		SELECT ` + postSelectCols + `
		FROM "post" p
		INNER JOIN "board" b ON b.id = p.board_id
		WHERE p.author_id = $1 AND ` + postReadableWhere(2, 3) + `
		ORDER BY p.create_time DESC
		LIMIT $4 OFFSET $5
	`
	err := db.Select(&list, sqlStr, profileUserID, postReaderUID(r), postReaderAdmin(r), limit, offset)
	return list, err
}

// CountCommentsByAuthorIDForViewer 统计评论数；所属帖须在浏览者视角可读。
func CountCommentsByAuthorIDForViewer(userID int64, r *PostReader) (int64, error) {
	var n int64
	err := db.Get(&n, `
		SELECT COUNT(*)
		FROM "comment" c
		INNER JOIN "post" p ON p.id = c.post_id
		INNER JOIN "board" b ON b.id = p.board_id
		WHERE c.author_id = $1 AND c.deleted_at IS NULL AND `+postReadableWhere(2, 3),
		userID, postReaderUID(r), postReaderAdmin(r),
	)
	return n, err
}

// ListCommentsByAuthorIDForViewer 评论列表；所属帖须在浏览者视角可读。
func ListCommentsByAuthorIDForViewer(userID int64, r *PostReader, limit, offset int) ([]models.UserHomeCommentItem, error) {
	var list []models.UserHomeCommentItem
	err := db.Select(&list, `
		SELECT
			c.id,
			c.post_id,
			p.title AS post_title,
			c.content,
			c.score,
			c.create_time,
			c.update_time
		FROM "comment" c
		INNER JOIN "post" p ON p.id = c.post_id
		INNER JOIN "board" b ON b.id = p.board_id
		WHERE c.author_id = $1 AND c.deleted_at IS NULL AND `+postReadableWhere(2, 3)+`
		ORDER BY c.create_time DESC
		LIMIT $4 OFFSET $5
	`, userID, postReaderUID(r), postReaderAdmin(r), limit, offset)
	return list, err
}
