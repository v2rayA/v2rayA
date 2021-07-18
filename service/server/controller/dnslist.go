package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/server/service"
)

func PutDnsList(ctx *gin.Context) {
	var data struct {
		Internal string `json:"internal"`
		External string `json:"external"`
	}
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		common.ResponseError(ctx, logError(nil, "bad request"))
		return
	}
	if len(data.Internal) == 0 && len(data.External) != 0 {
		common.ResponseError(ctx, logError(nil, "internal dns servers cannot be empty"))
		return
	}
	internal, err := service.RefineDnsList(data.Internal)
	if err != nil {
		common.ResponseError(ctx, logError(nil, fmt.Errorf("internal dns servers: %w", err)))
		return
	}
	external, err := service.RefineDnsList(data.External)
	if err != nil {
		common.ResponseError(ctx, logError(nil, fmt.Errorf("external dns servers: %w", err)))
		return
	}
	if err = configure.SetInternalDnsList(&internal); err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	if err = configure.SetExternalDnsList(&external); err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	common.ResponseSuccess(ctx, nil)
}

func GetDnsList(ctx *gin.Context) {
	common.ResponseSuccess(ctx, gin.H{
		"internal": configure.GetInternalDnsListNotNil(),
		"external": configure.GetExternalDnsListNotNil(),
	})
}
