package mysql

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"errors"
	"myreddit/models"
)

const secret = "hyanhsing0101"

var (
	ErrorUserExist       = errors.New("User Exist")
	ErrorUserNotExist    = errors.New("User Not Exist")
	ErrorInvalidPassword = errors.New("Wrong Password")
)

func CheckUserExist(username string) (err error) {
	sqlStr := "select count(user_id) from user where username = ?"
	var count int
	if err = db.Get(&count, sqlStr, username); err != nil {
		return err
	}
	if count > 0 {
		return ErrorUserExist
	}
	return
}

func Login(user *models.User) (err error) {
	oPassword := user.Password
	sqlStr := "select user_id, username, password from user where username = ?"
	err = db.Get(user, sqlStr, user.Username)
	if err == sql.ErrNoRows {
		return ErrorUserNotExist
	}
	if err != nil {
		return err
	}
	password := encryptPassword(oPassword, user.Username)
	if password != user.Password {
		return ErrorInvalidPassword
	}
	return
}

func InsertUser(user *models.User) error {
	// 对密码进行加密
	password := encryptPassword(user.Password, user.Username)
	// 执行 SQL 语句入库
	sqlStr := "insert into user (user_id, username, password) values (?, ?, ?)"
	if _, err := db.Exec(sqlStr, user.UserID, user.Username, password); err != nil {
		return err
	}
	return nil
}

func encryptPassword(opassword string, salt string) string {
	h := md5.New()
	h.Write([]byte(secret))
	h.Write([]byte(salt))
	h.Write([]byte(opassword))
	return hex.EncodeToString(h.Sum(nil))
}
