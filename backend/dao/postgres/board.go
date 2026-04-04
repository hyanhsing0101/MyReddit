package postgres

import (
	"database/sql"
	"errors"
	"myreddit/models"

	"github.com/lib/pq"
)

var (
	ErrorBoardNotExist  = errors.New("board not exist")
	ErrorBoardSlugTaken = errors.New("board slug taken")
)

func CountBoards(includeSystemSink bool) (int64, error) {
	var n int64
	sqlStr := `select count(*) from "board" where ($1 or is_system_sink = false)`
	err := db.Get(&n, sqlStr, includeSystemSink)
	return n, err
}

func ListBoards(includeSystemSink bool, limit, offset int) ([]models.Board, error) {
	var list []models.Board
	sqlStr := `
		select id, slug, name, description, created_by, is_system_sink, create_time, update_time
		from "board"
		where ($1 or is_system_sink = false)
		order by id asc
		limit $2 offset $3`
	err := db.Select(&list, sqlStr, includeSystemSink, limit, offset)
	return list, err
}

func GetBoardByID(id int64) (*models.Board, error) {
	var b models.Board
	sqlStr := `
		select id, slug, name, description, created_by, is_system_sink, create_time, update_time
		from "board"
		where id = $1`
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
		select id, slug, name, description, created_by, is_system_sink, create_time, update_time
		from "board"
		where slug = $1`
	err := db.Get(&b, sqlStr, slug)
	if err == sql.ErrNoRows {
		return nil, ErrorBoardNotExist
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func CreateBoard(b *models.Board) error {
	sqlStr := `
		insert into "board" (slug, name, description, created_by, is_system_sink, create_time, update_time)
		values ($1, $2, $3, $4, false, $5, $6)`
	_, err := db.Exec(sqlStr,
		b.Slug, b.Name, b.Description, b.CreatedBy,
		b.CreateTime, b.UpdateTime,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrorBoardSlugTaken
		}
		return err
	}
	return nil
}
