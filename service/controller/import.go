package controller

import (
	"V2RayA/service"
	"V2RayA/tools"
	"errors"
	"github.com/gin-gonic/gin"
)

func Import(ctx *gin.Context) {
	var data struct {
		URL string `json:"url"`
	}
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		tools.ResponseError(ctx, errors.New("参数有误"))
		return
	}
	err = service.Import(data.URL)
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	GetTouch(ctx)
}
