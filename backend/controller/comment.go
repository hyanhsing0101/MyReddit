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

func ListCommentsHandler(c *gin.Context) {
	idStr := c.Param("id")
	postID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || postID < 1 {
		ResponseError(c, CodeInvalidParam)
		return
	}
	p := new(models.ParamCommentList)
	if err := c.ShouldBindQuery(p); err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}
	data, err := logic.ListComments(postID, p, GetOptionalUserID(c))
	if err != nil {
		if errors.Is(err, postgres.ErrorPostNotExist) {
			ResponseError(c, CodePostNotExist)
			return
		}
		zap.L().Error("ListComments Failed", zap.Error(err), zap.Int64("post_id", postID))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, data)
}

func CreateCommentHandler(c *gin.Context) {
	userID, err := GetCurrentUser(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	idStr := c.Param("id")
	postID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || postID < 1 {
		ResponseError(c, CodeInvalidParam)
		return
	}
	p := new(models.ParamCreateComment)
	if err := c.ShouldBindJSON(&p); err != nil {
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			ResponseError(c, CodeInvalidParam)
			return
		}
		ResponseErrorWithMsg(c, CodeInvalidParam, errs.Translate(trans))
		return
	}
	if err := logic.CreateComment(postID, userID, p); err != nil {
		if errors.Is(err, postgres.ErrorPostNotExist) {
			ResponseError(c, CodePostNotExist)
			return
		}
		if errors.Is(err, postgres.ErrorCommentNotExist) {
			ResponseError(c, CodeCommentNotExist)
			return
		}
		if errors.Is(err, logic.ErrParentCommentMismatch) {
			ResponseError(c, CodeParentCommentMismatch)
			return
		}
		if errors.Is(err, logic.ErrInvalidCommentParent) {
			ResponseError(c, CodeInvalidCommentParent)
			return
		}
		if errors.Is(err, logic.ErrPostSealed) {
			ResponseError(c, CodePostSealed)
			return
		}
		if errors.Is(err, logic.ErrPostCommentsLocked) {
			ResponseError(c, CodePostCommentsLocked)
			return
		}
		zap.L().Error("CreateComment Failed", zap.Error(err), zap.Int64("post_id", postID))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseCreated(c, nil)
}
