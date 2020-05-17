package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/json-iterator/go"
	"time"
	"v2rayA/common"
	"v2rayA/persistence/configure"
	"v2rayA/service"
)

func GetPingLatency(ctx *gin.Context) {
	var wt []*configure.Which
	err := jsoniter.Unmarshal([]byte(ctx.Query("whiches")), &wt)
	if err != nil {
		common.ResponseError(ctx, logError(nil, "bad request"))
		return
	}
	wt, err = service.Ping(wt, 1*time.Second)
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	common.ResponseSuccess(ctx, gin.H{
		"whiches": wt,
	})
}

func GetHttpLatency(ctx *gin.Context) {
	var wt []*configure.Which
	err := jsoniter.Unmarshal([]byte(ctx.Query("whiches")), &wt)
	if err != nil {
		common.ResponseError(ctx, logError(nil, "bad request"))
		return
	}
	wt, err = service.TestHttpLatency(wt, 8*time.Second, 4, false)
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	common.ResponseSuccess(ctx, gin.H{
		"whiches": wt,
	})
}
