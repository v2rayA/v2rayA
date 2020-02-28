package controller

import (
	"V2RayA/persistence/configure"
	"V2RayA/service"
	"V2RayA/tools"
	"errors"
	"github.com/gin-gonic/gin"
)

func PostImport(ctx *gin.Context) {
	var data struct {
		URL   string `json:"url"`
		Which *configure.Which
	}
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		tools.ResponseError(ctx, errors.New("bad request"))
		return
	}
	err = service.Import(data.URL, data.Which)
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	GetTouch(ctx)
}
