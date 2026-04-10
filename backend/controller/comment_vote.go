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

func VoteCommentHandler(c *gin.Context) {
	userID, err := GetCurrentUser(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	postIDStr := c.Param("id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil || postID < 1 {
		ResponseError(c, CodeInvalidParam)
		return
	}
	cidStr := c.Param("cid")
	commentID, err := strconv.ParseInt(cidStr, 10, 64)
	if err != nil || commentID < 1 {
		ResponseError(c, CodeInvalidParam)
		return
	}
	p := new(models.ParamVotePost)
	if err := c.ShouldBindJSON(&p); err != nil {
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			ResponseError(c, CodeInvalidParam)
			return
		}
		ResponseErrorWithMsg(c, CodeInvalidParam, errs.Translate(trans))
		return
	}

	data, err := logic.VoteComment(postID, commentID, userID, *p.Value)
	if err != nil {
		if errors.Is(err, postgres.ErrorPostNotExist) {
			ResponseError(c, CodePostNotExist)
			return
		}
		if errors.Is(err, postgres.ErrorCommentNotExist) {
			ResponseErrorWithMsg(c, CodeInvalidParam, "comment not exist")
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
		zap.L().Error("VoteComment Failed", zap.Error(err), zap.Int64("post_id", postID), zap.Int64("comment_id", commentID))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, data)
}
