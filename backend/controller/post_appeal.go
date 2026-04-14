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

func UpsertMyPostAppealHandler(c *gin.Context) {
	userID, err := GetCurrentUser(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || postID < 1 {
		ResponseError(c, CodeInvalidParam)
		return
	}
	p := new(models.ParamCreatePostAppeal)
	if err := c.ShouldBindJSON(p); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			ResponseErrorWithMsg(c, CodeInvalidParam, errs.Translate(trans))
			return
		}
		ResponseError(c, CodeInvalidParam)
		return
	}
	if err := logic.UpsertMyPostAppeal(postID, userID, p); err != nil {
		switch {
		case errors.Is(err, postgres.ErrorPostNotExist):
			ResponseError(c, CodePostNotExist)
		case errors.Is(err, logic.ErrPostAppealForbidden):
			ResponseError(c, CodeForbidden)
		case errors.Is(err, logic.ErrCannotAppealUnsealedPost):
			ResponseError(c, CodeCannotAppealUnsealedPost)
		default:
			zap.L().Error("UpsertMyPostAppeal Failed", zap.Error(err), zap.Int64("post_id", postID), zap.Int64("user_id", userID))
			ResponseError(c, CodeServerBusy)
		}
		return
	}
	ResponseSuccess(c, nil)
}

func GetMyPostAppealHandler(c *gin.Context) {
	userID, err := GetCurrentUser(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || postID < 1 {
		ResponseError(c, CodeInvalidParam)
		return
	}
	data, err := logic.GetMyPostAppeal(postID, userID)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrorPostNotExist):
			ResponseError(c, CodePostNotExist)
		case errors.Is(err, postgres.ErrorPostAppealNotExist):
			ResponseError(c, CodePostAppealNotExist)
		case errors.Is(err, logic.ErrPostAppealForbidden):
			ResponseError(c, CodeForbidden)
		default:
			zap.L().Error("GetMyPostAppeal Failed", zap.Error(err), zap.Int64("post_id", postID), zap.Int64("user_id", userID))
			ResponseError(c, CodeServerBusy)
		}
		return
	}
	ResponseSuccess(c, data)
}

func ListBoardPostAppealsHandler(c *gin.Context) {
	userID, err := GetCurrentUser(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	boardID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || boardID < 1 {
		ResponseError(c, CodeInvalidParam)
		return
	}
	p := new(models.ParamPostAppealList)
	if err := c.ShouldBindQuery(p); err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}
	data, err := logic.ListBoardPostAppeals(boardID, userID, p)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrorBoardNotExist):
			ResponseError(c, CodeBoardNotExist)
		case errors.Is(err, logic.ErrManageReportForbidden):
			ResponseError(c, CodeForbidden)
		default:
			zap.L().Error("ListBoardPostAppeals Failed", zap.Error(err), zap.Int64("board_id", boardID), zap.Int64("user_id", userID))
			ResponseError(c, CodeServerBusy)
		}
		return
	}
	ResponseSuccess(c, data)
}

func HandlePostAppealHandler(c *gin.Context) {
	userID, err := GetCurrentUser(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	appealID, err := strconv.ParseInt(c.Param("aid"), 10, 64)
	if err != nil || appealID < 1 {
		ResponseError(c, CodeInvalidParam)
		return
	}
	p := new(models.ParamHandlePostAppeal)
	if err := c.ShouldBindJSON(p); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			ResponseErrorWithMsg(c, CodeInvalidParam, errs.Translate(trans))
			return
		}
		ResponseError(c, CodeInvalidParam)
		return
	}
	if err := logic.HandlePostAppeal(appealID, userID, p); err != nil {
		switch {
		case errors.Is(err, postgres.ErrorPostAppealNotExist):
			ResponseError(c, CodePostAppealNotExist)
		case errors.Is(err, logic.ErrManageReportForbidden):
			ResponseError(c, CodeForbidden)
		default:
			zap.L().Error("HandlePostAppeal Failed", zap.Error(err), zap.Int64("appeal_id", appealID), zap.Int64("user_id", userID))
			ResponseError(c, CodeServerBusy)
		}
		return
	}
	ResponseSuccess(c, nil)
}
