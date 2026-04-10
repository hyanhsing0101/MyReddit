package postgres

import (
	"database/sql"
	"errors"
	"myreddit/models"
	"time"
)

var ErrorPostNotExist = errors.New("post not exist")

const postSelectCols = `p.id, p.board_id, p.title, p.content, p.author_id, p.deleted_at,
		p.sealed_at, p.sealed_by_user_id, p.seal_kind, p.score, p.create_time, p.update_time,
		b.slug AS board_slug, b.name AS board_name, b.visibility AS board_visibility`

func postReadableWhere(uidIdx, adminIdx int) string {
	return `p.deleted_at IS NULL AND ` + sqlPostBoardReadable(uidIdx, adminIdx) + ` AND ` + sqlPostSealReadable(uidIdx, adminIdx)
}

// CreatePost 写入帖子并返回新帖 id。
func CreatePost(post *models.Post) (int64, error) {
	sqlStr := `
		INSERT INTO "post" (board_id, title, content, author_id, create_time, update_time)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id;
		`
	var id int64
	err := db.Get(&id, sqlStr, post.BoardID, post.Title, post.Content, post.AuthorID, post.CreateTime, post.UpdateTime)
	if err != nil {
		return 0, err
	}
	return id, err
}

// CountPosts 统计读者可见的未软删帖子（私有板与封帖按 PostReader 过滤）。
func CountPosts(boardID *int64, r *PostReader) (int64, error) {
	var n int64
	var err error
	uid := postReaderUID(r)
	adm := postReaderAdmin(r)
	base := `
		SELECT COUNT(*) FROM "post" p
		INNER JOIN "board" b ON b.id = p.board_id
		WHERE ` + postReadableWhere(1, 2)
	if boardID == nil {
		err = db.Get(&n, base, uid, adm)
	} else {
		err = db.Get(&n, base+` AND p.board_id = $3`, uid, adm, *boardID)
	}
	return n, err
}

// ListPosts 按排序规则分页查询读者可见帖子。
func ListPosts(boardID *int64, sort models.PostSort, limit, offset int, r *PostReader) ([]models.Post, error) {
	var list []models.Post
	var err error
	orderBy := postOrderBy(sort)
	uid := postReaderUID(r)
	adm := postReaderAdmin(r)
	base := `
		SELECT ` + postSelectCols + `
		FROM "post" p
		INNER JOIN "board" b ON b.id = p.board_id
		WHERE ` + postReadableWhere(1, 2)

	if boardID == nil {
		sqlStr := base + ` ORDER BY ` + orderBy + ` LIMIT $3 OFFSET $4`
		err = db.Select(&list, sqlStr, uid, adm, limit, offset)
	} else {
		sqlStr := base + ` AND p.board_id = $3 ORDER BY ` + orderBy + ` LIMIT $4 OFFSET $5`
		err = db.Select(&list, sqlStr, uid, adm, *boardID, limit, offset)
	}
	return list, err
}

// ListPostsByIDsOrdered 按给定 ids 顺序返回读者可见帖子（过滤已软删及不可见）。
func ListPostsByIDsOrdered(ids []int64, r *PostReader) ([]models.Post, error) {
	if len(ids) == 0 {
		return []models.Post{}, nil
	}
	uid := postReaderUID(r)
	adm := postReaderAdmin(r)
	sqlStr := `
		SELECT ` + postSelectCols + `
		FROM unnest($1::bigint[]) WITH ORDINALITY AS x(id, ord)
		JOIN "post" p ON p.id = x.id
		JOIN "board" b ON b.id = p.board_id
		WHERE ` + postReadableWhere(2, 3) + `
		ORDER BY x.ord;
	`
	var list []models.Post
	if err := db.Select(&list, sqlStr, Int64Array(ids), uid, adm); err != nil {
		return nil, err
	}
	return list, nil
}

func postOrderBy(sort models.PostSort) string {
	switch sort {
	case models.PostSortTop:
		return "p.score DESC, p.create_time DESC"
	case models.PostSortHot:
		return "((GREATEST(p.score, -50))::double precision / POWER(GREATEST(EXTRACT(EPOCH FROM (NOW() - p.create_time)) / 3600.0, 0) + 2.0, 1.8)) DESC, p.create_time DESC"
	default:
		return "p.create_time DESC"
	}
}

// GetPostByID 读者视角：未软删且通过私有板/封帖规则。
func GetPostByID(id int64, r *PostReader) (*models.Post, error) {
	var p models.Post
	sqlStr := `
		SELECT ` + postSelectCols + `
		FROM "post" p
		INNER JOIN "board" b ON b.id = p.board_id
		WHERE p.id = $1 AND ` + postReadableWhere(2, 3)
	err := db.Get(&p, sqlStr, id, postReaderUID(r), postReaderAdmin(r))
	if err == sql.ErrNoRows {
		return nil, ErrorPostNotExist
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// GetPostByIDIncludingDeleted 用于删帖鉴权等（含已软删）；不含读者过滤。
func GetPostByIDIncludingDeleted(id int64) (*models.Post, error) {
	var p models.Post
	sqlStr := `
		SELECT ` + postSelectCols + `
		FROM "post" p
		INNER JOIN "board" b ON b.id = p.board_id
		WHERE p.id = $1`
	err := db.Get(&p, sqlStr, id)
	if err == sql.ErrNoRows {
		return nil, ErrorPostNotExist
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// SoftDeletePost 软删帖子并更新更新时间；帖子不存在时返回 ErrorPostNotExist。
func SoftDeletePost(id int64, at time.Time) error {
	res, err := db.Exec(
		`UPDATE "post" SET deleted_at = $2, update_time = $2 WHERE id = $1 AND deleted_at IS NULL`,
		id, at,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrorPostNotExist
	}
	return nil
}

// UpdatePostContent 更新帖子标题/正文与更新时间；帖子不存在时返回 ErrorPostNotExist。
func UpdatePostContent(id int64, title, content string, at time.Time) error {
	res, err := db.Exec(
		`UPDATE "post" SET title = $2, content = $3, update_time = $4 WHERE id = $1 AND deleted_at IS NULL`,
		id, title, content, at,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrorPostNotExist
	}
	return nil
}

// SealPost 封帖（与软删独立）；已删或已封返回 ErrorPostNotExist。
func SealPost(postID int64, sealedByUserID int64, sealKind string, at time.Time) error {
	res, err := db.Exec(`
		UPDATE "post" SET sealed_at = $2, sealed_by_user_id = $3, seal_kind = $4, update_time = $2
		WHERE id = $1 AND deleted_at IS NULL AND sealed_at IS NULL`,
		postID, at, sealedByUserID, sealKind,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrorPostNotExist
	}
	return nil
}

// UnsealPost 解封；未封返回 ErrorPostNotExist。
func UnsealPost(postID int64, at time.Time) error {
	res, err := db.Exec(`
		UPDATE "post" SET sealed_at = NULL, sealed_by_user_id = NULL, seal_kind = NULL, update_time = $2
		WHERE id = $1 AND sealed_at IS NOT NULL`,
		postID, at,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrorPostNotExist
	}
	return nil
}
