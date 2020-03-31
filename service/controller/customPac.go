package controller

import (
	"V2RayA/common"
	"V2RayA/core/v2ray/asset"
	"V2RayA/persistence/configure"
	"V2RayA/service"
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
