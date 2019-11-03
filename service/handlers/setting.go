package handlers

import (
	"V2RayA/global"
	"V2RayA/models/touch"
	"V2RayA/tools"
	"errors"
	"github.com/gin-gonic/gin"
	"os"
)

func GetSetting(ctx *gin.Context) {
	tr := global.GetTouchRaw()
	tr.Lock()
	defer tr.Unlock()
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
	switch data.PacMode {
	case touch.GfwlistMode:
		if _, err := os.Stat("/etc/v2ray/h2y.dat"); err != nil {
			tools.ResponseError(ctx, errors.New("未发现GFWList文件，请更新GFWList后再试"))
			return
		}
	case touch.CustomMode:
		if _, err := os.Stat("/etc/v2ray/custom.dat"); err != nil {
			tools.ResponseError(ctx, errors.New("未发现custom.dat文件，功能正在开发"))
			return
		}
	}

	tr := global.GetTouchRaw()
	tr.Lock()
	defer tr.Unlock()
	tr.Setting = &data
	global.SetTouchRaw(&tr)
	err = tr.WriteToFile()
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	//重写配置并重启连接，使得对PAC的修改立即生效
	if tr.ConnectedServer != nil {
		tsr, _ := tr.LocateServer(tr.ConnectedServer)
		err = tools.UpdateV2RayConfigAndRestart(&tsr.VmessInfo)
		if err != nil {
			tools.ResponseError(ctx, err)
			return
		}
	}
	tools.ResponseSuccess(ctx, nil)
}
