package controller

import (
	"V2RayA/persistence/configure"
	"V2RayA/service"
	"V2RayA/tools"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"time"
)

func GetPingLatency(ctx *gin.Context) {
	var wt []configure.Which
	err := json.Unmarshal([]byte(ctx.Query("whiches")), &wt)
	if err != nil {
		tools.ResponseError(ctx, errors.New("参数有误"))
		return
	}
	wt, err = service.Ping(wt, 5, 5*time.Second)
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	tools.ResponseSuccess(ctx, gin.H{
		"whiches": wt,
	})
}
