package controller

import (
	"time"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/server/service"
)

func GetNodeLatency(ctx *gin.Context) {
	getNodeLatency(ctx, true)
}

func GetPingLatency(ctx *gin.Context) {
	getNodeLatency(ctx, false)
}

func getNodeLatency(ctx *gin.Context, configured bool) {
	release, ok := beginMutation(ctx)
	if !ok {
		return
	}
	defer release()

	var wt []*configure.Which
	err := jsoniter.Unmarshal([]byte(ctx.Query("whiches")), &wt)
	if err != nil {
		common.ResponseError(ctx, badRequest("whiches", "query parameter \"whiches\" must be a JSON array of server items"))
		return
	}
	setting := configure.GetSettingNotNil()
	if configured && setting.NodeProbeMode == "proxy" {
		wt, err = service.TestHttpLatency(ctx.Request.Context(), wt, 8*time.Second, 32, false, setting.NodeProbeURL)
	} else {
		wt, err = service.Ping(wt, time.Second)
	}
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	common.ResponseSuccess(ctx, gin.H{
		"whiches": wt,
	})
}

func GetHttpLatency(ctx *gin.Context) {
	release, ok := beginMutation(ctx)
	if !ok {
		return
	}
	defer release()

	var wt []*configure.Which
	err := jsoniter.Unmarshal([]byte(ctx.Query("whiches")), &wt)
	if err != nil {
		common.ResponseError(ctx, badRequest("whiches", "query parameter \"whiches\" must be a JSON array of server items"))
		return
	}
	wt, err = service.TestHttpLatency(ctx.Request.Context(), wt, 8*time.Second, 32, false, ctx.Query("testUrl"))
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	common.ResponseSuccess(ctx, gin.H{
		"whiches": wt,
	})
}
