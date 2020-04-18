package controller

import (
	"v2rayA/common"
	"v2rayA/persistence/configure"
	"v2rayA/service"
	"github.com/gin-gonic/gin"
	"github.com/json-iterator/go"
)

func GetSharingAddress(ctx *gin.Context) {
	var w configure.Which
	err := jsoniter.Unmarshal([]byte(ctx.Query("touch")), &w)
	if err != nil {
		common.ResponseError(ctx, logError(nil, "bad request"))
		return
	}
	addr, err := service.GetSharingAddress(&w)
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	common.ResponseSuccess(ctx, gin.H{"sharingAddress": addr})
}
