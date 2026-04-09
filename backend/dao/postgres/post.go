package postgres

import (
	"database/sql"
	"errors"
	"myreddit/models"
	"time"
)

var ErrorPostNotExist = errors.New("post not exist")

const postSelectCols = `p.id, p.board_id, p.title, p.content, p.author_id, p.deleted_at, p.score, p.create_time, p.update_time,
		       b.slug as board_slug, b.name as board_name`

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

func CountPosts(boardID *int64) (int64, error) {
	var n int64
	var err error
	active := ` and deleted_at is null`
	if boardID == nil {
		err = db.Get(&n, `select count(*) from "post" where deleted_at is null`)
	} else {
		err = db.Get(&n, `select count(*) from "post" where board_id = $1`+active, *boardID)
	}
	return n, err
}

func ListPosts(boardID *int64, sort models.PostSort, limit, offset int) ([]models.Post, error) {
	var list []models.Post
	var err error
	base := `
		select ` + postSelectCols + `
		from "post" p
		inner join "board" b on b.id = p.board_id
		where p.deleted_at is null`
	orderBy := postOrderBy(sort)

	if boardID == nil {
		sqlStr := base + ` order by ` + orderBy + ` limit $1 offset $2`
		err = db.Select(&list, sqlStr, limit, offset)
	} else {
		sqlStr := base + ` and p.board_id = $1 order by ` + orderBy + ` limit $2 offset $3`
		err = db.Select(&list, sqlStr, *boardID, limit, offset)
	}
	return list, err
}

// ListPostsByIDsOrdered 按给定 ids 的顺序返回帖子（过滤已软删）。
func ListPostsByIDsOrdered(ids []int64) ([]models.Post, error) {
	if len(ids) == 0 {
		return []models.Post{}, nil
	}
	sqlStr := `
		SELECT ` + postSelectCols + `
		FROM unnest($1::bigint[]) WITH ORDINALITY AS x(id, ord)
		JOIN "post" p ON p.id = x.id
		JOIN "board" b ON b.id = p.board_id
		WHERE p.deleted_at IS NULL
		ORDER BY x.ord;
	`
	var list []models.Post
	if err := db.Select(&list, sqlStr, Int64Array(ids)); err != nil {
		return nil, err
	}
	return list, nil
}

func postOrderBy(sort models.PostSort) string {
	switch sort {
	case models.PostSortTop:
		return "p.score DESC, p.create_time DESC"
	case models.PostSortHot:
		// 简单热度：分数 / 时间衰减（小时）
		// 用 GREATEST 防止除以 0；2.0 是启动保护，避免新帖瞬时过高
		return "((GREATEST(p.score, -50))::double precision / POWER(GREATEST(EXTRACT(EPOCH FROM (NOW() - p.create_time)) / 3600.0, 0) + 2.0, 1.8)) DESC, p.create_time DESC"
	default:
		return "p.create_time DESC"
	}
}

// GetPostByID 仅未软删的帖子（列表/详情对所有人一致：已删即不存在）
func GetPostByID(id int64) (*models.Post, error) {
	var p models.Post
	sqlStr := `
		select ` + postSelectCols + `
		from "post" p
		inner join "board" b on b.id = p.board_id
		where p.id = $1 and p.deleted_at is null`
	err := db.Get(&p, sqlStr, id)
	if err == sql.ErrNoRows {
		return nil, ErrorPostNotExist
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// GetPostByIDIncludingDeleted 用于删帖鉴权（含已软删行）
func GetPostByIDIncludingDeleted(id int64) (*models.Post, error) {
	var p models.Post
	sqlStr := `
		select ` + postSelectCols + `
		from "post" p
		inner join "board" b on b.id = p.board_id
		where p.id = $1`
	err := db.Get(&p, sqlStr, id)
	if err == sql.ErrNoRows {
		return nil, ErrorPostNotExist
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func SoftDeletePost(id int64, at time.Time) error {
	res, err := db.Exec(
		`update "post" set deleted_at = $2, update_time = $2 where id = $1 and deleted_at is null`,
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

func UpdatePostContent(id int64, title, content string, at time.Time) error {
	res, err := db.Exec(
		`update "post" set title = $2, content = $3, update_time = $4 where id = $1 and deleted_at is null`,
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
