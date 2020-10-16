package controller

import (
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/core/v2ray/asset"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/service"
	"github.com/gin-gonic/gin"
)

func GetSiteDatFiles(ctx *gin.Context) {
	common.ResponseSuccess(ctx, gin.H{"siteDatFiles": service.GetSiteDatFiles()})
}
func GetCustomPac(ctx *gin.Context) {
	common.ResponseSuccess(ctx, gin.H{
		"customPac":          configure.GetCustomPacNotNil(),
		"V2RayLocationAsset": asset.GetV2rayLocationAsset(),
	})
}
func PutCustomPac(ctx *gin.Context) {
	var data struct {
		CustomPac configure.CustomPac `json:"customPac"`
	}
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		common.ResponseError(ctx, logError(err, "bad request"))
		return
	}
	err = configure.SetCustomPac(&data.CustomPac)
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	common.ResponseSuccess(ctx, nil)
}
