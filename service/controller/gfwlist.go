package controller

import (
	"V2RayA/core/gfwlist"
	"V2RayA/common"
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