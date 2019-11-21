package controller

import (
	"V2RayA/persistence/configure"
	"V2RayA/service"
	"V2RayA/tools"
	"github.com/json-iterator/go"
	"errors"
	"github.com/gin-gonic/gin"
)

func GetSharingAddress(ctx *gin.Context) {
	var w configure.Which
	err := jsoniter.Unmarshal([]byte(ctx.Query("touch")), &w)
	if err != nil {
		tools.ResponseError(ctx, errors.New("参数有误"))
		return
	}
	addr, err := service.GetSharingAddress(&w)
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	tools.ResponseSuccess(ctx, gin.H{"sharingAddress": addr})
}
