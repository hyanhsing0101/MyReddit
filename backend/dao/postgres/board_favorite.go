package postgres

import (
	"database/sql"
	"myreddit/models"
	"time"
)

func AddBoardFavorite(userID, boardID int64) error {
	_, err := db.Exec(`
		INSERT INTO board_favorite (user_id, board_id, create_time)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, board_id) DO NOTHING
	`, userID, boardID, time.Now())
	return err
}

// HasBoardFavorite 用户是否已收藏（订阅）该板块。
func HasBoardFavorite(userID, boardID int64) (bool, error) {
	var n int64
	err := db.Get(&n, `
		SELECT COUNT(*) FROM board_favorite WHERE user_id = $1 AND board_id = $2`,
		userID, boardID,
	)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func RemoveBoardFavorite(userID, boardID int64) error {
	_, err := db.Exec(`
		DELETE FROM board_favorite WHERE user_id = $1 AND board_id = $2
	`, userID, boardID)
	return err
}

func CountBoardFavoritesByUser(userID int64) (int64, error) {
	var n int64
	err := db.Get(&n, `SELECT COUNT(*) FROM board_favorite WHERE user_id = $1`, userID)
	return n, err
}

// ListBoardIDsFavoritedByUser 在给定 board_ids 中返回该用户已收藏的 id 集合。
func ListBoardIDsFavoritedByUser(userID int64, boardIDs []int64) (map[int64]struct{}, error) {
	if len(boardIDs) == 0 {
		return map[int64]struct{}{}, nil
	}
	var favIDs []int64
	err := db.Select(&favIDs, `
		SELECT board_id FROM board_favorite
		WHERE user_id = $1 AND board_id = ANY($2)
	`, userID, Int64Array(boardIDs))
	if err != nil {
		return nil, err
	}
	out := make(map[int64]struct{}, len(favIDs))
	for _, id := range favIDs {
		out[id] = struct{}{}
	}
	return out, nil
}

// ListBoardFavoritesByUser 按收藏时间倒序
func ListBoardFavoritesByUser(userID int64, limit, offset int) ([]models.Board, []time.Time, error) {
	type row struct {
		ID           int64          `db:"id"`
		Slug         string         `db:"slug"`
		Name         string         `db:"name"`
		Description  sql.NullString `db:"description"`
		CreatedBy    sql.NullInt64  `db:"created_by"`
		Visibility   string         `db:"visibility"`
		IsSystemSink bool           `db:"is_system_sink"`
		CreateTime   time.Time      `db:"create_time"`
		UpdateTime   time.Time      `db:"update_time"`
		FavoritedAt  time.Time      `db:"favorited_at"`
	}
	var rows []row
	err := db.Select(&rows, `
		SELECT b.id, b.slug, b.name, b.description, b.created_by, b.visibility, b.is_system_sink, b.create_time, b.update_time,
		       bf.create_time AS favorited_at
		FROM board_favorite bf
		INNER JOIN "board" b ON b.id = bf.board_id
		WHERE bf.user_id = $1
		ORDER BY bf.create_time DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, nil, err
	}
	boards := make([]models.Board, len(rows))
	times := make([]time.Time, len(rows))
	for i, r := range rows {
		boards[i] = models.Board{
			ID:           r.ID,
			Slug:         r.Slug,
			Name:         r.Name,
			Description:  r.Description,
			CreatedBy:    r.CreatedBy,
			Visibility:   r.Visibility,
			IsSystemSink: r.IsSystemSink,
			CreateTime:   r.CreateTime,
			UpdateTime:   r.UpdateTime,
		}
		times[i] = r.FavoritedAt
	}
	return boards, times, nil
}
