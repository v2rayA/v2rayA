package controller

import (
	"V2RayA/common"
	"V2RayA/core/v2ray/asset/gfwlist"
	"github.com/gin-gonic/gin"
)

func PutGFWList(ctx *gin.Context) {
	localGFWListVersion, err := gfwlist.CheckAndUpdateGFWList()
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	common.ResponseSuccess(ctx, gin.H{
		"localGFWListVersion": localGFWListVersion,
	})
}