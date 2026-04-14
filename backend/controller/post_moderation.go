package controller

import (
	"errors"
	"myreddit/dao/postgres"
	"myreddit/logic"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func LockPostCommentsHandler(c *gin.Context) {
	updatePostModerationState(c, logic.LockPostComments, "LockPostComments")
}

func UnlockPostCommentsHandler(c *gin.Context) {
	updatePostModerationState(c, logic.UnlockPostComments, "UnlockPostComments")
}

func PinPostHandler(c *gin.Context) {
	updatePostModerationState(c, logic.PinPost, "PinPost")
}

func UnpinPostHandler(c *gin.Context) {
	updatePostModerationState(c, logic.UnpinPost, "UnpinPost")
}

func updatePostModerationState(c *gin.Context, fn func(postID, operatorID int64) error, op string) {
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
	if err := fn(id, userID); err != nil {
		if errors.Is(err, postgres.ErrorPostNotExist) {
			ResponseError(c, CodePostNotExist)
			return
		}
		if errors.Is(err, logic.ErrLockPostForbidden) ||
			errors.Is(err, logic.ErrUnlockPostForbidden) ||
			errors.Is(err, logic.ErrPinPostForbidden) ||
			errors.Is(err, logic.ErrUnpinPostForbidden) {
			ResponseError(c, CodeForbidden)
			return
		}
		zap.L().Error(op+" Failed", zap.Error(err), zap.Int64("post_id", id))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, nil)
}
