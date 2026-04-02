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
