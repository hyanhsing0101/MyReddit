package controller

import (
	"errors"
	"myreddit/dao/postgres"
	"myreddit/logic"
	"myreddit/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ListBoardModerationLogsHandler(c *gin.Context) {
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
	p := new(models.ParamModerationLogList)
	if err := c.ShouldBindQuery(p); err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}
	data, err := logic.ListBoardModerationLogs(boardID, userID, p)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrorBoardNotExist):
			ResponseError(c, CodeBoardNotExist)
		case errors.Is(err, logic.ErrManageReportForbidden):
			ResponseError(c, CodeForbidden)
		default:
			zap.L().Error("ListBoardModerationLogs Failed", zap.Error(err), zap.Int64("board_id", boardID), zap.Int64("user_id", userID))
			ResponseError(c, CodeServerBusy)
		}
		return
	}
	ResponseSuccess(c, data)
}

func GetModerationDashboardHandler(c *gin.Context) {
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
	data, err := logic.GetModerationDashboard(boardID, userID)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrorBoardNotExist):
			ResponseError(c, CodeBoardNotExist)
		case errors.Is(err, logic.ErrManageReportForbidden):
			ResponseError(c, CodeForbidden)
		default:
			zap.L().Error("GetModerationDashboard Failed", zap.Error(err), zap.Int64("board_id", boardID), zap.Int64("user_id", userID))
			ResponseError(c, CodeServerBusy)
		}
		return
	}
	ResponseSuccess(c, data)
}
