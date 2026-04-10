package postgres

import "myreddit/models"

// CountPostsByAuthorID 统计某用户未软删帖子数量。
func CountPostsByAuthorID(userID int64) (int64, error) {
	var n int64
	err := db.Get(&n, `
		SELECT COUNT(*)
		FROM "post"
		WHERE author_id = $1 AND deleted_at IS NULL
	`, userID)
	return n, err
}

// ListPostsByAuthorID 按创建时间倒序分页查询某用户帖子。
func ListPostsByAuthorID(userID int64, limit, offset int) ([]models.Post, error) {
	var list []models.Post
	sqlStr := `
		SELECT ` + postSelectCols + `
		FROM "post" p
		INNER JOIN "board" b ON b.id = p.board_id
		WHERE p.author_id = $1 AND p.deleted_at IS NULL
		ORDER BY p.create_time DESC
		LIMIT $2 OFFSET $3
	`
	err := db.Select(&list, sqlStr, userID, limit, offset)
	return list, err
}

// CountCommentsByAuthorID 统计某用户未软删评论数量。
func CountCommentsByAuthorID(userID int64) (int64, error) {
	var n int64
	err := db.Get(&n, `
		SELECT COUNT(*)
		FROM "comment"
		WHERE author_id = $1 AND deleted_at IS NULL
	`, userID)
	return n, err
}

// ListCommentsByAuthorID 按创建时间倒序分页查询某用户评论，并附带所属帖子标题。
func ListCommentsByAuthorID(userID int64, limit, offset int) ([]models.UserHomeCommentItem, error) {
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
		WHERE c.author_id = $1 AND c.deleted_at IS NULL AND p.deleted_at IS NULL
		ORDER BY c.create_time DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	return list, err
}
