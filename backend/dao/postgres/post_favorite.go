package postgres

import (
	"myreddit/models"
	"time"

	"github.com/lib/pq"
)

func AddPostFavorite(userID, postID int64) error {
	_, err := db.Exec(`
		INSERT INTO post_favorite (user_id, post_id, create_time)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, post_id) DO NOTHING
	`, userID, postID, time.Now())
	return err
}

func RemovePostFavorite(userID, postID int64) error {
	_, err := db.Exec(`
		DELETE FROM post_favorite WHERE user_id = $1 AND post_id = $2
	`, userID, postID)
	return err
}

func CountPostFavoritesByUser(userID int64) (int64, error) {
	var n int64
	err := db.Get(&n, `
		SELECT COUNT(*)
		FROM post_favorite pf
		INNER JOIN "post" p ON p.id = pf.post_id
		WHERE pf.user_id = $1 AND p.deleted_at IS NULL
	`, userID)
	return n, err
}

// ListPostIDsFavoritedByUser 在给定 post_ids 中返回该用户已收藏的 id 集合。
func ListPostIDsFavoritedByUser(userID int64, postIDs []int64) (map[int64]struct{}, error) {
	if len(postIDs) == 0 {
		return map[int64]struct{}{}, nil
	}
	var favIDs []int64
	err := db.Select(&favIDs, `
		SELECT post_id FROM post_favorite
		WHERE user_id = $1 AND post_id = ANY($2)
	`, userID, pq.Array(postIDs))
	if err != nil {
		return nil, err
	}
	out := make(map[int64]struct{}, len(favIDs))
	for _, id := range favIDs {
		out[id] = struct{}{}
	}
	return out, nil
}

// ListPostFavoritesByUser 按收藏时间倒序；已软删帖子不返回。
func ListPostFavoritesByUser(userID int64, limit, offset int) ([]models.Post, []time.Time, error) {
	type row struct {
		models.Post
		FavoritedAt time.Time `db:"favorited_at"`
	}
	var rows []row
	err := db.Select(&rows, `
		SELECT `+postSelectCols+`,
		       pf.create_time AS favorited_at
		FROM post_favorite pf
		INNER JOIN "post" p ON p.id = pf.post_id
		INNER JOIN "board" b ON b.id = p.board_id
		WHERE pf.user_id = $1 AND p.deleted_at IS NULL
		ORDER BY pf.create_time DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, nil, err
	}
	posts := make([]models.Post, len(rows))
	times := make([]time.Time, len(rows))
	for i, r := range rows {
		posts[i] = r.Post
		times[i] = r.FavoritedAt
	}
	return posts, times, nil
}
