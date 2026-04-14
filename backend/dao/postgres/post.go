package postgres

import (
	"database/sql"
	"errors"
	"myreddit/models"
	"time"
)

var ErrorPostNotExist = errors.New("post not exist")

const postSelectCols = `p.id, p.board_id, p.title, p.content, p.author_id, p.deleted_at,
		p.sealed_at, p.sealed_by_user_id, p.seal_kind,
		p.comments_locked_at, p.comments_locked_by_user_id,
		p.pinned_at, p.pinned_by_user_id,
		p.score, p.create_time, p.update_time,
		b.slug AS board_slug, b.name AS board_name, b.visibility AS board_visibility`

func postReadableWhere(uidIdx, adminIdx int) string {
	return `p.deleted_at IS NULL AND ` + sqlPostBoardReadable(uidIdx, adminIdx) + ` AND ` + sqlPostSealReadable(uidIdx, adminIdx)
}

// sqlSubscribedBoardFilter 限定帖子所在板块在用户「收藏板块」列表中（$1 须与读者 user id 一致）。
func sqlSubscribedBoardFilter() string {
	return ` AND EXISTS (
		SELECT 1 FROM board_favorite bf WHERE bf.board_id = p.board_id AND bf.user_id = $1)`
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
// subscribedOnly 为 true 时仅统计用户已收藏板块（board_favorite）下的帖子；须与登录读者 r.UserID 一致（由上层保证）。
func CountPosts(boardID *int64, subscribedOnly bool, r *PostReader) (int64, error) {
	var n int64
	var err error
	uid := postReaderUID(r)
	adm := postReaderAdmin(r)
	base := `
		SELECT COUNT(*) FROM "post" p
		INNER JOIN "board" b ON b.id = p.board_id
		WHERE ` + postReadableWhere(1, 2)
	sub := ""
	if subscribedOnly {
		sub = sqlSubscribedBoardFilter()
	}
	if boardID == nil {
		err = db.Get(&n, base+sub, uid, adm)
	} else {
		err = db.Get(&n, base+sub+` AND p.board_id = $3`, uid, adm, *boardID)
	}
	return n, err
}

// ListPosts 按排序规则分页查询读者可见帖子。
func ListPosts(boardID *int64, subscribedOnly bool, sort models.PostSort, limit, offset int, r *PostReader) ([]models.Post, error) {
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
	sub := ""
	if subscribedOnly {
		sub = sqlSubscribedBoardFilter()
	}

	if boardID == nil {
		sqlStr := base + sub + ` ORDER BY ` + orderBy + ` LIMIT $3 OFFSET $4`
		err = db.Select(&list, sqlStr, uid, adm, limit, offset)
	} else {
		sqlStr := base + sub + ` AND p.board_id = $3 ORDER BY ` + orderBy + ` LIMIT $4 OFFSET $5`
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
	// 板块治理：置顶帖优先展示，再按各排序规则。
	pinned := "CASE WHEN p.pinned_at IS NULL THEN 1 ELSE 0 END ASC, p.pinned_at DESC, "
	switch sort {
	case models.PostSortTop:
		return pinned + "p.score DESC, p.create_time DESC"
	case models.PostSortHot:
		return pinned + "((GREATEST(p.score, -50))::double precision / POWER(GREATEST(EXTRACT(EPOCH FROM (NOW() - p.create_time)) / 3600.0, 0) + 2.0, 1.8)) DESC, p.create_time DESC"
	default:
		return pinned + "p.create_time DESC"
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

// RestorePost 取消软删（仅当 deleted_at 非空时生效）。
func RestorePost(id int64, at time.Time) error {
	res, err := db.Exec(
		`UPDATE "post" SET deleted_at = NULL, update_time = $2 WHERE id = $1 AND deleted_at IS NOT NULL`,
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

// CountBoardDeletedPosts 本板已软删帖子数（版主治理列表用）。
func CountBoardDeletedPosts(boardID int64) (int64, error) {
	var n int64
	err := db.Get(&n, `SELECT COUNT(*) FROM "post" WHERE board_id = $1 AND deleted_at IS NOT NULL`, boardID)
	return n, err
}

// ListBoardDeletedPosts 本板已软删帖子，按删除时间倒序。
func ListBoardDeletedPosts(boardID int64, limit, offset int) ([]models.DeletedPostView, error) {
	type row struct {
		ID        int64         `db:"id"`
		Title     string        `db:"title"`
		AuthorID  sql.NullInt64 `db:"author_id"`
		DeletedAt time.Time     `db:"deleted_at"`
	}
	var rows []row
	err := db.Select(&rows, `
		SELECT p.id, p.title, p.author_id, p.deleted_at
		FROM "post" p
		WHERE p.board_id = $1 AND p.deleted_at IS NOT NULL
		ORDER BY p.deleted_at DESC
		LIMIT $2 OFFSET $3`,
		boardID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	out := make([]models.DeletedPostView, len(rows))
	for i, r := range rows {
		v := models.DeletedPostView{ID: r.ID, Title: r.Title, DeletedAt: r.DeletedAt}
		if r.AuthorID.Valid {
			x := r.AuthorID.Int64
			v.AuthorID = &x
		}
		out[i] = v
	}
	return out, nil
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

// LockPostComments 锁帖评论（禁止新评论）；已删或已锁返回 ErrorPostNotExist。
func LockPostComments(postID, operatorID int64, at time.Time) error {
	res, err := db.Exec(`
		UPDATE "post" SET comments_locked_at = $2, comments_locked_by_user_id = $3, update_time = $2
		WHERE id = $1 AND deleted_at IS NULL AND comments_locked_at IS NULL`,
		postID, at, operatorID,
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

// UnlockPostComments 解锁评论；未锁返回 ErrorPostNotExist。
func UnlockPostComments(postID int64, at time.Time) error {
	res, err := db.Exec(`
		UPDATE "post" SET comments_locked_at = NULL, comments_locked_by_user_id = NULL, update_time = $2
		WHERE id = $1 AND comments_locked_at IS NOT NULL`,
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

// PinPost 置顶帖子；已删或已置顶返回 ErrorPostNotExist。
func PinPost(postID, operatorID int64, at time.Time) error {
	res, err := db.Exec(`
		UPDATE "post" SET pinned_at = $2, pinned_by_user_id = $3, update_time = $2
		WHERE id = $1 AND deleted_at IS NULL AND pinned_at IS NULL`,
		postID, at, operatorID,
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

// UnpinPost 取消置顶；未置顶返回 ErrorPostNotExist。
func UnpinPost(postID int64, at time.Time) error {
	res, err := db.Exec(`
		UPDATE "post" SET pinned_at = NULL, pinned_by_user_id = NULL, update_time = $2
		WHERE id = $1 AND pinned_at IS NOT NULL`,
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
