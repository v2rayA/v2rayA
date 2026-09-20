package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db/configure"
)

func GetDomainsExcluded(ctx *gin.Context) {
	common.ResponseSuccess(ctx, gin.H{
		"domains": configure.GetDomainsExcluded(),
	})
}

func PutDomainsExcluded(ctx *gin.Context) {
	release, ok := beginMutation(ctx)
	if !ok {
		return
	}
	defer release()
	var data struct {
		DomainList string `json:"domains"`
	}
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		common.ResponseError(ctx, badRequest("domains", "request body must be {\"domains\": string}"))
		return
	}
	configure.SetDomainsExcluded(data.DomainList)
	common.ResponseSuccess(ctx, nil)
}
