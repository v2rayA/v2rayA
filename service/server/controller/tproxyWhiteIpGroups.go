package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db/configure"
)

func GetTproxyWhiteIpGroups(ctx *gin.Context) {
	resp := configure.GetTproxyWhiteIpGroups()
	common.ResponseSuccess(ctx, gin.H{
		"countryCodes": resp.CountryCodes,
		"customIps":    resp.CustomIps,
	})
}

func PutTproxyWhiteIpGroups(ctx *gin.Context) {
	var data struct {
		CountryCodes []string `json:"countryCodes"`
		CustomIps    []string `json:"customIps"`
	}
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		common.ResponseError(ctx, logError("bad request"))
		return
	}
	configure.SetTproxyWhiteIpGroups(data.CountryCodes, data.CustomIps)
	common.ResponseSuccess(ctx, nil)
}
