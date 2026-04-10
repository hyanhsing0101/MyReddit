package controller

import (
	"myreddit/logic"
	"myreddit/models"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func SearchHandler(c *gin.Context) {
	p := new(models.ParamSearch)
	if err := c.ShouldBindQuery(p); err != nil {
		zap.L().Error("Search With Invalid Param", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}
	searchData, err := logic.Search(p, GetOptionalUserID(c))
	if err != nil {
		zap.L().Error("Search With Error", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, searchData)
}