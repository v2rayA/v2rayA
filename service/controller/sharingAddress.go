package controller

import (
	"V2RayA/persistence/configure"
	"V2RayA/service"
	"V2RayA/common"
	"github.com/json-iterator/go"
	"errors"
	"github.com/gin-gonic/gin"
)

func GetSharingAddress(ctx *gin.Context) {
	var w configure.Which
	err := jsoniter.Unmarshal([]byte(ctx.Query("touch")), &w)
	if err != nil {
		common.ResponseError(ctx, errors.New("bad request"))
		return
	}
	addr, err := service.GetSharingAddress(&w)
	if err != nil {
		common.ResponseError(ctx, err)
		return
	}
	common.ResponseSuccess(ctx, gin.H{"sharingAddress": addr})
}
