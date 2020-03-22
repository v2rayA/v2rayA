package controller

import (
	"V2RayA/core/v2ray/asset"
	"V2RayA/persistence/configure"
	"V2RayA/service"
	"V2RayA/tools"
	"errors"
	"github.com/gin-gonic/gin"
)

func GetSetting(ctx *gin.Context) {
	s := service.GetSetting()
	var localGFWListVersion string
	t, err := asset.GetGFWListModTime()
	if err == nil {
		localGFWListVersion = t.Local().Format("2006-01-02")
	}
	tools.ResponseSuccess(ctx, gin.H{
		"setting":              s,
		"localGFWListVersion":  localGFWListVersion,
	})
}

func PutSetting(ctx *gin.Context) {
	var data configure.Setting
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		tools.ResponseError(ctx, errors.New("bad request"+err.Error()))
		return
	}
	if data.MuxOn == configure.Yes && (data.Mux < 1 || data.Mux > 1024) {
		tools.ResponseError(ctx, errors.New("mux should be between 1 and 1024"))
		return
	}
	err = service.UpdateSetting(&data)
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	tools.ResponseSuccess(ctx, nil)
}
