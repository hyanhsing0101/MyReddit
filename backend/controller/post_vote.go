package controller

import (
	"errors"
	"myreddit/dao/postgres"
	"myreddit/logic"
	"myreddit/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

func VotePostHandler(c *gin.Context) {
	userID, err := GetCurrentUser(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	idStr := c.Param("id")
	postID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || postID < 1 {
		zap.L().Error("Vote Post With Invalid Param", zap.String("id", idStr))
		ResponseError(c, CodeInvalidParam)
		return
	}
	p := new(models.ParamVotePost)
	if err := c.ShouldBindJSON(&p); err != nil {
		zap.L().Error("Vote Post With Invalid JSON", zap.Error(err))
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			ResponseError(c, CodeInvalidParam)
			return
		}
		ResponseErrorWithMsg(c, CodeInvalidParam, errs.Translate(trans))
		return
	}

	data, err := logic.VotePost(postID, userID, *p.Value)
	if err != nil {
		if errors.Is(err, postgres.ErrorPostNotExist) {
			ResponseError(c, CodePostNotExist)
			return
		}
		if errors.Is(err, postgres.ErrorInvalidVoteValue) {
			ResponseErrorWithMsg(c, CodeInvalidParam, "value must be 1, -1, or 0")
			return
		}
		if errors.Is(err, logic.ErrPostSealed) {
			ResponseError(c, CodePostSealed)
			return
		}
		zap.L().Error("Vote Post Failed", zap.Error(err), zap.Int64("post_id", postID))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, data)
}
