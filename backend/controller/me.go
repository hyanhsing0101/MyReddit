package controller

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"myreddit/logic"
)

func MePermissionsHandler(c *gin.Context) {
	userID, err := GetCurrentUser(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	data, err := logic.MePermissions(userID)
	if err != nil {
		zap.L().Error("MePermissions", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, data)
}
