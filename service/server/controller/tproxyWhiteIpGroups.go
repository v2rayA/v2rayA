package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db/configure"
)

func GetTproxyWhiteIpGroups(ctx *gin.Context) {
	common.ResponseSuccess(ctx, gin.H{
		"list": configure.GetTproxyWhiteIpGroups(),
	})
}

func PutTproxyWhiteIpGroups(ctx *gin.Context) {
	var data struct {
		List []string `json:"list"`
	}
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		common.ResponseError(ctx, logError("bad request"))
		return
	}
	configure.SetTproxyWhiteIpGroups(data.List)
	common.ResponseSuccess(ctx, nil)
}
