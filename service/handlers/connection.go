package handlers

import (
	"V2RayA/global"
	"V2RayA/models/touch"
	"V2RayA/tools"
	"errors"
	"github.com/gin-gonic/gin"
)

func PostConnection(ctx *gin.Context) {
	var data touch.WhichTouch
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		tools.ResponseError(ctx, errors.New("参数有误"))
		return
	}
	//定位Server
	tr := global.GetTouchRaw()
	lastConnectedServer := tr.ConnectedServer
	tsr, err := tr.LocateServer(&data)
	if err != nil {
		tools.ResponseError(ctx, errors.New("无效的参数"))
		return
	}
	//根据找到的Server更新V2Ray服务的配置
	err = tools.UpdateV2RayConfig(&tsr.VmessInfo)
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	//保存节点连接成功的结果
	tr.Lock()
	defer tr.Unlock()
	tr.SetDisConnect()   //断连现有连接
	tsr.Connected = true //tsr是个指针
	tr.ConnectedServer = &data
	err = tr.WriteToFile()
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	global.SetTouchRaw(&tr)
	err = tools.EnableV2rayService()
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	tools.ResponseSuccess(ctx, gin.H{"connectedServer": tr.ConnectedServer, "lastConnectedServer": lastConnectedServer})
}

func DeleteConnection(ctx *gin.Context) {
	tr := global.GetTouchRaw()
	cs := tr.ConnectedServer
	tr.Lock() //写操作加锁
	defer tr.Unlock()
	tr.SetDisConnect()
	err := tr.WriteToFile()
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	err = tools.StopV2rayService()
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	err = tools.DisableV2rayService()
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	global.SetTouchRaw(&tr)
	tools.ResponseSuccess(ctx, gin.H{"lastConnectedServer": cs})
}
