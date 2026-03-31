package mysql

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"myreddit/models"
)

const secret = "hyanhsing0101"

func CheckUserExist(username string) (err error) {
	sqlStr := "select count(user_id) from user where username = ?"
	var count int
	if err = db.Get(&count, sqlStr, username); err != nil {
		return err
	}
	if count > 0 {
		return errors.New("username exist")
	}
	return
}

func QueryUserByUsername() {

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
