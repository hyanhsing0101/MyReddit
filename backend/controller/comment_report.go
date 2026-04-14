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

func CreateCommentReportHandler(c *gin.Context) {
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
	commentID, err := strconv.ParseInt(c.Param("cid"), 10, 64)
	if err != nil || commentID < 1 {
		ResponseError(c, CodeInvalidParam)
		return
	}
	p := new(models.ParamCreatePostReport)
	if err := c.ShouldBindJSON(p); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			ResponseErrorWithMsg(c, CodeInvalidParam, errs.Translate(trans))
			return
		}
		ResponseError(c, CodeInvalidParam)
		return
	}
	if err := logic.CreateCommentReport(postID, commentID, userID, p); err != nil {
		switch {
		case errors.Is(err, postgres.ErrorPostNotExist):
			ResponseError(c, CodePostNotExist)
		case errors.Is(err, postgres.ErrorCommentNotExist):
			ResponseError(c, CodeCommentNotExist)
		case errors.Is(err, logic.ErrCannotReportOwnComment):
			ResponseError(c, CodeCannotReportOwnComment)
		case errors.Is(err, logic.ErrDuplicateActiveCommentReport):
			ResponseError(c, CodeDuplicateCommentReport)
		default:
			zap.L().Error("CreateCommentReport Failed", zap.Error(err), zap.Int64("post_id", postID), zap.Int64("comment_id", commentID), zap.Int64("user_id", userID))
			ResponseError(c, CodeServerBusy)
		}
		return
	}
	ResponseCreated(c, nil)
}

func ListBoardCommentReportsHandler(c *gin.Context) {
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
	p := new(models.ParamReportList)
	if err := c.ShouldBindQuery(p); err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}
	data, err := logic.ListBoardCommentReports(boardID, userID, p)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrorBoardNotExist):
			ResponseError(c, CodeBoardNotExist)
		case errors.Is(err, logic.ErrManageReportForbidden):
			ResponseError(c, CodeForbidden)
		default:
			zap.L().Error("ListBoardCommentReports Failed", zap.Error(err), zap.Int64("board_id", boardID), zap.Int64("user_id", userID))
			ResponseError(c, CodeServerBusy)
		}
		return
	}
	ResponseSuccess(c, data)
}

func UpdateCommentReportStatusHandler(c *gin.Context) {
	userID, err := GetCurrentUser(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	reportID, err := strconv.ParseInt(c.Param("rid"), 10, 64)
	if err != nil || reportID < 1 {
		ResponseError(c, CodeInvalidParam)
		return
	}
	p := new(models.ParamUpdatePostReport)
	if err := c.ShouldBindJSON(p); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			ResponseErrorWithMsg(c, CodeInvalidParam, errs.Translate(trans))
			return
		}
		ResponseError(c, CodeInvalidParam)
		return
	}
	if err := logic.UpdateCommentReportStatus(reportID, userID, p); err != nil {
		switch {
		case errors.Is(err, postgres.ErrorCommentReportNotExist):
			ResponseError(c, CodeCommentReportNotExist)
		case errors.Is(err, logic.ErrManageReportForbidden):
			ResponseError(c, CodeForbidden)
		default:
			zap.L().Error("UpdateCommentReportStatus Failed", zap.Error(err), zap.Int64("report_id", reportID), zap.Int64("user_id", userID))
			ResponseError(c, CodeServerBusy)
		}
		return
	}
	ResponseSuccess(c, nil)
}

func BatchUpdateCommentReportStatusHandler(c *gin.Context) {
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
	p := new(models.ParamBatchUpdateCommentReports)
	if err := c.ShouldBindJSON(p); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			ResponseErrorWithMsg(c, CodeInvalidParam, errs.Translate(trans))
			return
		}
		ResponseError(c, CodeInvalidParam)
		return
	}
	if err := logic.BatchUpdateCommentReportStatus(boardID, userID, p); err != nil {
		switch {
		case errors.Is(err, postgres.ErrorBoardNotExist):
			ResponseError(c, CodeBoardNotExist)
		case errors.Is(err, postgres.ErrorCommentReportNotExist):
			ResponseError(c, CodeCommentReportNotExist)
		case errors.Is(err, logic.ErrManageReportForbidden):
			ResponseError(c, CodeForbidden)
		default:
			zap.L().Error("BatchUpdateCommentReportStatus Failed", zap.Error(err), zap.Int64("board_id", boardID), zap.Int64("user_id", userID))
			ResponseError(c, CodeServerBusy)
		}
		return
	}
	ResponseSuccess(c, nil)
}
