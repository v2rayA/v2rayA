package controller

import (
	"github.com/mzz2017/v2rayA/common"
	"github.com/mzz2017/v2rayA/core/v2ray/asset/gfwlist"
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