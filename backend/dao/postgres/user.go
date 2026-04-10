package postgres

import (
	"database/sql"
	"errors"
)

var (
	ErrorUserExist    = errors.New("User Exist")
	ErrorUserNotExist = errors.New("User Not Exist")
)

func CheckUserExist(username string) (err error) {
	sqlStr := `select count(user_id) from "user" where username = $1`
	var count int
	if err = db.Get(&count, sqlStr, username); err != nil {
		return err
	}
	if count > 0 {
		return ErrorUserExist
	}
	return
}

func CheckUserByIDAndName(userID int64, username string) (bool, error) {
	sqlStr := `select count(user_id) from "user" where user_id = $1 and username = $2`
	var count int
	if err := db.Get(&count, sqlStr, userID, username); err != nil {
		return false, err
	}
	return count == 1, nil
}

// UserAuthRow 仅用于登录鉴权所需字段读取。
type UserAuthRow struct {
	UserID   int64  `db:"user_id"`
	Username string `db:"username"`
	Password string `db:"password"`
}

func GetUserAuthByUsername(username string) (*UserAuthRow, error) {
	var row UserAuthRow
	sqlStr := `select user_id, username, password from "user" where username = $1`
	err := db.Get(&row, sqlStr, username)
	if err == sql.ErrNoRows {
		return nil, ErrorUserNotExist
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func InsertUser(userID int64, username, passwordHash string) error {
	sqlStr := `insert into "user" (user_id, username, password) values ($1, $2, $3)`
	if _, err := db.Exec(sqlStr, userID, username, passwordHash); err != nil {
		return err
	}
	return nil
}

// IsSiteAdmin 判断用户是否为站长
func IsSiteAdmin(userID int64) (bool, error) {
	var admin bool
	sqlStr := `select coalesce(is_site_admin, false) from "user" where user_id = $1`
	err := db.Get(&admin, sqlStr, userID)
	if err != nil {
		return false, err
	}
	return admin, nil
}

// GetUsernameByUserID 获取用户名
func GetUsernameByUserID(userID int64) (string, error) {
	var username string
	sqlStr := `select username from "user" where user_id = $1`
	err := db.Get(&username, sqlStr, userID)
	if err == sql.ErrNoRows {
		return "", ErrorUserNotExist
	}
	if err != nil {
		return "", err
	}
	return username, nil
}
