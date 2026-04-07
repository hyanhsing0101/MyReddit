package controller

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

const ContextUserIDkey = "UserID"

var ErrorUserNotLogin = fmt.Errorf("user not login")

func GetCurrentUser(c *gin.Context) (userID int64, err error) {
	uid, ok := c.Get(ContextUserIDkey)
	if !ok {
		err = ErrorUserNotLogin
		return
	}
	userID, ok = uid.(int64)
	if !ok {
		err = ErrorUserNotLogin
		return
	}
	return
}

// GetOptionalUserID 未登录或未设置 UserID 时返回 nil。
func GetOptionalUserID(c *gin.Context) *int64 {
	v, ok := c.Get(ContextUserIDkey)
	if !ok {
		return nil
	}
	id, ok := v.(int64)
	if !ok {
		return nil
	}
	return &id
}
