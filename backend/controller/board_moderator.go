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

func ListBoardModeratorsHandler(c *gin.Context) {
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
	list, err := logic.ListBoardModerators(boardID, userID)
	if err != nil {
		if errors.Is(err, postgres.ErrorBoardNotExist) {
			ResponseError(c, CodeBoardNotExist)
			return
		}
		zap.L().Error("ListBoardModerators Failed", zap.Error(err), zap.Int64("board_id", boardID))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, gin.H{"list": list})
}

func AddBoardModeratorHandler(c *gin.Context) {
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
	p := new(models.ParamAddBoardModerator)
	if err := c.ShouldBindJSON(p); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			ResponseErrorWithMsg(c, CodeInvalidParam, errs.Translate(trans))
			return
		}
		ResponseError(c, CodeInvalidParam)
		return
	}
	if p.UserID < 1 {
		ResponseError(c, CodeInvalidParam)
		return
	}
	if err := logic.AddOrUpdateBoardModerator(boardID, userID, p); err != nil {
		switch {
		case errors.Is(err, postgres.ErrorBoardNotExist):
			ResponseError(c, CodeBoardNotExist)
		case errors.Is(err, postgres.ErrorUserNotExist):
			ResponseError(c, CodeUserNotExist)
		case errors.Is(err, logic.ErrManageBoardModeratorsForbidden):
			ResponseError(c, CodeForbidden)
		default:
			zap.L().Error("AddBoardModerator Failed", zap.Error(err), zap.Int64("board_id", boardID))
			ResponseError(c, CodeServerBusy)
		}
		return
	}
	ResponseSuccess(c, nil)
}

func RemoveBoardModeratorHandler(c *gin.Context) {
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
	targetUserID, err := strconv.ParseInt(c.Param("uid"), 10, 64)
	if err != nil || targetUserID < 1 {
		ResponseError(c, CodeInvalidParam)
		return
	}
	if err := logic.RemoveBoardModerator(boardID, targetUserID, userID); err != nil {
		switch {
		case errors.Is(err, postgres.ErrorBoardNotExist):
			ResponseError(c, CodeBoardNotExist)
		case errors.Is(err, logic.ErrBoardModeratorNotExist):
			ResponseError(c, CodeForbidden)
		case errors.Is(err, logic.ErrManageBoardModeratorsForbidden),
			errors.Is(err, logic.ErrCannotRemoveLastOwner):
			ResponseError(c, CodeForbidden)
		default:
			zap.L().Error("RemoveBoardModerator Failed", zap.Error(err), zap.Int64("board_id", boardID), zap.Int64("target_user_id", targetUserID))
			ResponseError(c, CodeServerBusy)
		}
		return
	}
	ResponseSuccess(c, nil)
}

func UpdateBoardModeratorRoleHandler(c *gin.Context) {
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
	targetUserID, err := strconv.ParseInt(c.Param("uid"), 10, 64)
	if err != nil || targetUserID < 1 {
		ResponseError(c, CodeInvalidParam)
		return
	}
	p := new(models.ParamUpdateBoardModeratorRole)
	if err := c.ShouldBindJSON(p); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			ResponseErrorWithMsg(c, CodeInvalidParam, errs.Translate(trans))
			return
		}
		ResponseError(c, CodeInvalidParam)
		return
	}
	if err := logic.UpdateBoardModeratorRole(boardID, targetUserID, userID, p); err != nil {
		switch {
		case errors.Is(err, postgres.ErrorBoardNotExist):
			ResponseError(c, CodeBoardNotExist)
		case errors.Is(err, logic.ErrBoardModeratorNotExist):
			ResponseError(c, CodeForbidden)
		case errors.Is(err, logic.ErrManageBoardModeratorsForbidden),
			errors.Is(err, logic.ErrCannotRemoveLastOwner):
			ResponseError(c, CodeForbidden)
		default:
			zap.L().Error("UpdateBoardModeratorRole Failed", zap.Error(err), zap.Int64("board_id", boardID), zap.Int64("target_user_id", targetUserID))
			ResponseError(c, CodeServerBusy)
		}
		return
	}
	ResponseSuccess(c, nil)
}
