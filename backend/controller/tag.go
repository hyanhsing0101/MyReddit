package controller

import (
	"myreddit/logic"
	"myreddit/models"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ListTagsHandler(c *gin.Context) {
	p := new(models.ParamTagList)
	if err := c.ShouldBindQuery(p); err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}
	data, err := logic.ListTags(p)
	if err != nil {
		zap.L().Error("ListTags failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, data)
}