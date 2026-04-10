package postgres

import (
	"database/sql"
	"errors"
	"myreddit/models"
	"strconv"
)

var (
	ErrorBoardNotExist  = errors.New("board not exist")
	ErrorBoardSlugTaken = errors.New("board slug taken")
)

// sqlBoardListVisible 板块列表可见：公开、系统板、或当前用户为成员/版主/站主。
func sqlBoardListVisible(viewerArgIdx, adminArgIdx int) string {
	u := strconv.Itoa(viewerArgIdx)
	a := strconv.Itoa(adminArgIdx)
	return `(
	b.is_system_sink = TRUE
	OR b.visibility = 'public'
	OR ($` + u + `::bigint IS NOT NULL AND EXISTS (
		SELECT 1 FROM board_favorite bf WHERE bf.board_id = b.id AND bf.user_id = $` + u + `))
	OR ($` + u + `::bigint IS NOT NULL AND EXISTS (
		SELECT 1 FROM board_moderator bm WHERE bm.board_id = b.id AND bm.user_id = $` + u + `))
	OR $` + a + ` = TRUE
)`
}

// CountBoards 统计板块数；viewer 用于过滤私有板（匿名仅见公开板与系统板）。
func CountBoards(includeSystemSink bool, viewerID *int64, viewerIsSiteAdmin bool) (int64, error) {
	var n int64
	var uid interface{}
	if viewerID != nil {
		uid = *viewerID
	} else {
		uid = nil
	}
	q := `
		SELECT COUNT(*) FROM "board" b
		WHERE ($1 OR b.is_system_sink = FALSE) AND ` + sqlBoardListVisible(2, 3)
	err := db.Get(&n, q, includeSystemSink, uid, viewerIsSiteAdmin)
	return n, err
}

// ListBoards 分页；顺序 id asc 与旧行为一致。
func ListBoards(includeSystemSink bool, limit, offset int, viewerID *int64, viewerIsSiteAdmin bool) ([]models.Board, error) {
	var list []models.Board
	var uid interface{}
	if viewerID != nil {
		uid = *viewerID
	} else {
		uid = nil
	}
	q := `
		SELECT b.id, b.slug, b.name, b.description, b.created_by, b.visibility, b.is_system_sink, b.create_time, b.update_time
		FROM "board" b
		WHERE ($1 OR b.is_system_sink = FALSE) AND ` + sqlBoardListVisible(4, 5) + `
		ORDER BY b.id ASC
		LIMIT $2 OFFSET $3`
	err := db.Select(&list, q, includeSystemSink, limit, offset, uid, viewerIsSiteAdmin)
	return list, err
}

func GetBoardByID(id int64) (*models.Board, error) {
	var b models.Board
	sqlStr := `
		SELECT id, slug, name, description, created_by, visibility, is_system_sink, create_time, update_time
		FROM "board"
		WHERE id = $1`
	err := db.Get(&b, sqlStr, id)
	if err == sql.ErrNoRows {
		return nil, ErrorBoardNotExist
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func GetBoardBySlug(slug string) (*models.Board, error) {
	var b models.Board
	sqlStr := `
		SELECT id, slug, name, description, created_by, visibility, is_system_sink, create_time, update_time
		FROM "board"
		WHERE slug = $1`
	err := db.Get(&b, sqlStr, slug)
	if err == sql.ErrNoRows {
		return nil, ErrorBoardNotExist
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// CreateBoardWithModerator 创建板块并将创建者写入版主表（同一事务）。
func CreateBoardWithModerator(b *models.Board, moderatorUserID int64) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	vis := b.Visibility
	if vis == "" {
		vis = models.BoardVisibilityPublic
	}

	const ins = `
		INSERT INTO "board" (slug, name, description, created_by, visibility, is_system_sink, create_time, update_time)
		VALUES ($1, $2, $3, $4, $5, false, $6, $7)
		RETURNING id`
	var id int64
	err = tx.Get(&id, ins, b.Slug, b.Name, b.Description, b.CreatedBy, vis, b.CreateTime, b.UpdateTime)
	if err != nil {
		if IsUniqueViolation(err) {
			return ErrorBoardSlugTaken
		}
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO board_moderator (user_id, board_id, role, appointed_by, create_time, update_time)
		VALUES ($1, $2, $3, $1, $4, $4)
		ON CONFLICT (user_id, board_id) DO NOTHING`,
		moderatorUserID, id, models.BoardModeratorRoleOwner, b.CreateTime,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}
