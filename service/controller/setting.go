package controller

import (
	"V2RayA/model/v2ray"
	"V2RayA/persistence/configure"
	"V2RayA/service"
	"V2RayA/tools"
	"errors"
	"github.com/gin-gonic/gin"
)

func GetSetting(ctx *gin.Context) {
	s := service.GetSetting()
	var localGFWListVersion, customPacFileVersion string
	t, err := v2ray.GetH2yModTime()
	if err == nil {
		localGFWListVersion = t.Format("2006-01-02")
	}
	t, err = v2ray.GetCustomModTime()
	if err == nil {
		customPacFileVersion = t.Format("2006-01-02")
	}
	tools.ResponseSuccess(ctx, gin.H{
		"setting":              s,
		"localGFWListVersion":  localGFWListVersion,
		"customPacFileVersion": customPacFileVersion,
	})
}

func PutSetting(ctx *gin.Context) {
	var data configure.Setting
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		tools.ResponseError(ctx, errors.New("参数有误"+err.Error()))
		return
	}
	if data.MuxOn == configure.Yes && (data.Mux < 1 || data.Mux > 1024) {
		tools.ResponseError(ctx, errors.New("多路复用最大并发连接数必须介于1到1024之间"))
		return
	}
	err = service.UpdateSetting(&data)
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	tools.ResponseSuccess(ctx, nil)
}
