package controller

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/kernel/v2ray/asset/dat"
)

func PutGFWList(ctx *gin.Context) {
	var data struct {
		DownloadLink string `json:"downloadLink"`
	}
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		common.ResponseError(ctx, badRequest("downloadLink", "request body must be {\"downloadLink\": string}"))
		return
	}
	localGFWListVersion, err := dat.CheckAndUpdateGFWList(data.DownloadLink)
	if errors.Is(err, dat.ErrGFWListUpToDate) {
		// Nothing to download is a result, not a failure.
		common.ResponseSuccess(ctx, gin.H{
			"localGFWListVersion": localGFWListVersion,
			"alreadyUpToDate":     true,
		})
		return
	}
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	common.ResponseSuccess(ctx, gin.H{
		"localGFWListVersion": localGFWListVersion,
		"alreadyUpToDate":     false,
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
