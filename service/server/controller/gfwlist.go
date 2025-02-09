package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/core/v2ray/asset/dat"
)

func PutGFWList(ctx *gin.Context) {
	var data struct {
		DownloadLink string `json:"downloadLink"`
	}
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		common.ResponseError(ctx, logError("bad request"))
		return
	}
	localGFWListVersion, err := dat.CheckAndUpdateGFWList(data.DownloadLink)
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	common.ResponseSuccess(ctx, gin.H{
		"localGFWListVersion": localGFWListVersion,
	})
}

func DeleteGFWList(ctx *gin.Context) {
	err := dat.DeleteGFWList()
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	common.ResponseSuccess(ctx, gin.H{})
}
