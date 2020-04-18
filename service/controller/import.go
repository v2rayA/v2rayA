package controller

import (
	"v2rayA/common"
	"v2rayA/persistence/configure"
	"v2rayA/service"
	"github.com/gin-gonic/gin"
)

func PostImport(ctx *gin.Context) {
	var data struct {
		URL   string `json:"url"`
		Which *configure.Which
	}
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		common.ResponseError(ctx, logError(nil, "bad request"))
		return
	}
	err = service.Import(data.URL, data.Which)
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	GetTouch(ctx)
}
