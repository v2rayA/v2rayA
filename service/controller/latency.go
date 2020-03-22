package controller

import (
	"V2RayA/persistence/configure"
	"V2RayA/service"
	"V2RayA/common"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/json-iterator/go"
	"time"
)

func GetPingLatency(ctx *gin.Context) {
	var wt []configure.Which
	err := jsoniter.Unmarshal([]byte(ctx.Query("whiches")), &wt)
	if err != nil {
		common.ResponseError(ctx, errors.New("bad request"))
		return
	}
	wt, err = service.Ping(wt, 5*time.Second)
	if err != nil {
		common.ResponseError(ctx, err)
		return
	}
	common.ResponseSuccess(ctx, gin.H{
		"whiches": wt,
	})
}

func GetHttpLatency(ctx *gin.Context) {
	var wt []configure.Which
	err := jsoniter.Unmarshal([]byte(ctx.Query("whiches")), &wt)
	if err != nil {
		common.ResponseError(ctx, errors.New("bad request"))
		return
	}
	wt, err = service.TestHttpLatency(wt, 10*time.Second, 4)
	if err != nil {
		common.ResponseError(ctx, err)
		return
	}
	common.ResponseSuccess(ctx, gin.H{
		"whiches": wt,
	})
}
