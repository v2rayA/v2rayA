package controller

import (
	"V2RayA/core/gfwlist"
	"V2RayA/tools"
	"github.com/gin-gonic/gin"
)

func PutGFWList(ctx *gin.Context) {
	localGFWListVersion, err := gfwlist.CheckAndUpdateGFWList()
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	tools.ResponseSuccess(ctx, gin.H{
		"localGFWListVersion": localGFWListVersion,
	})
}