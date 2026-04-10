package controller

import (
	"errors"
	"myreddit/dao/postgres"
	"myreddit/logic"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SealPostHandler 版主或站主封帖。
func SealPostHandler(c *gin.Context) {
	userID, err := GetCurrentUser(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		ResponseError(c, CodeInvalidParam)
		return
	}
	if err := logic.SealPost(id, userID); err != nil {
		if errors.Is(err, postgres.ErrorPostNotExist) {
			ResponseError(c, CodePostNotExist)
			return
		}
		if errors.Is(err, logic.ErrSealPostForbidden) {
			ResponseError(c, CodeForbidden)
			return
		}
		zap.L().Error("SealPost Failed", zap.Error(err), zap.Int64("post_id", id))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, nil)
}

// UnsealPostHandler 按规则解封（站主任意；版主仅 moderator 封帖）。
func UnsealPostHandler(c *gin.Context) {
	userID, err := GetCurrentUser(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		ResponseError(c, CodeInvalidParam)
		return
	}
	if err := logic.UnsealPost(id, userID); err != nil {
		if errors.Is(err, postgres.ErrorPostNotExist) {
			ResponseError(c, CodePostNotExist)
			return
		}
		if errors.Is(err, logic.ErrUnsealPostForbidden) {
			ResponseError(c, CodeForbidden)
			return
		}
		zap.L().Error("UnsealPost Failed", zap.Error(err), zap.Int64("post_id", id))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, nil)
}
