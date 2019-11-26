package controller

import (
	"V2RayA/persistence/configure"
	"V2RayA/service"
	"V2RayA/tools"
	"errors"
	"github.com/gin-gonic/gin"
)

func PutPorts(ctx *gin.Context) {
	var data configure.Ports
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		tools.ResponseError(ctx, errors.New("参数有误"))
		return
	}
	err = service.SetPorts(&data)
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	tools.ResponseSuccess(ctx, nil)
}

func GetPorts(ctx *gin.Context) {
	tools.ResponseSuccess(ctx, service.GetPorts())
}
