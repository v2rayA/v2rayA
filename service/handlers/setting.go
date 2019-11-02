package handlers

import (
	"V2RayA/global"
	"V2RayA/models/touch"
	"V2RayA/tools"
	"errors"
	"github.com/gin-gonic/gin"
)

func GetSetting(ctx *gin.Context) {
	tr := global.GetTouchRaw()
	tr.Lock()
	s := tr.Setting
	if s == nil {
		s = touch.NewSetting()
	}
	var localGFWListVersion, customPacFileVersion string
	t, err := tools.GetFileModTime("/etc/v2ray/h2y.dat")
	if err == nil {
		localGFWListVersion = t.Format("2006-01-02")
	}
	t, err = tools.GetFileModTime("/etc/v2ray/custom.dat")
	if err == nil {
		customPacFileVersion = t.Format("2006-01-02")
	}
	global.SetTouchRaw(&tr)
	err = tr.WriteToFile()
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	tr.Unlock()
	tools.ResponseSuccess(ctx, gin.H{
		"setting":              s,
		"localGFWListVersion":  localGFWListVersion,
		"customPacFileVersion": customPacFileVersion,
	})
}

func PutSetting(ctx *gin.Context) {
	var data touch.Setting
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		tools.ResponseError(ctx, errors.New("参数有误"))
		return
	}

	//TODO: 检查参数合法性

	tr := global.GetTouchRaw()
	tr.Lock()
	tr.Setting = &data
	global.SetTouchRaw(&tr)
	err = tr.WriteToFile()
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	tr.Unlock()
	tools.ResponseSuccess(ctx, nil)
}
