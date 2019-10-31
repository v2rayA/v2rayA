package handlers

import (
	"V2RayA/config"
	"V2RayA/models"
	"V2RayA/tools"
	"errors"
	"github.com/gin-gonic/gin"
)

func GetTouch(ctx *gin.Context) {
	tr := config.GetTouchRaw()
	running := tools.IsV2RayRunning()
	if !running {
		tr.SetDisConnect() //如果没有运行，就不给前端返回结点连接信息了
	}
	t := tr.ToTouch() //读操作就不锁了
	for i := range t.Subscriptions {
		t.Subscriptions[i].TYPE = models.SubscriptionType
		for j := range t.Subscriptions[i].Servers {
			t.Subscriptions[i].Servers[j].TYPE = models.SubscriptionServerType
		}
	}
	for i := range t.Servers {
		t.Servers[i].TYPE = models.ServerType
	}
	tools.ResponseSuccess(ctx, gin.H{
		"running": running,
		"touch":   t,
	})
}

func DeleteTouch(ctx *gin.Context) {
	var data models.WhichTouches
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		tools.ResponseError(ctx, errors.New("参数有误"))
		return
	}
	tr := config.GetTouchRaw()
	//对要删除的touch去重
	data.SetTouches(data.GetNonDuplicatedTouches(&tr))
	//对要删除的touch排序，将大的下标排在前面，从后往前删
	data.Sort()
	touches := data.GetTouches()
	tr.Lock() //写操作需要上锁
	defer tr.Unlock()
	disconnect := func() {
		tr.SetDisConnect()
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
	}
	for _, v := range touches {
		ind := v.ID - 1
		switch v.TYPE {
		case models.SubscriptionType: //这里删的是某个订阅
			//检查现在连接的结点是否在该订阅中，是的话断开连接
			if tr.ConnectedServer != nil && tr.ConnectedServer.TYPE == models.SubscriptionServerType && tr.ConnectedServer.Sub == ind {
				disconnect()
			}
			tr.Subscriptions = append(tr.Subscriptions[:ind], tr.Subscriptions[ind+1:]...)
		case models.ServerType:
			tr.Servers = append(tr.Servers[:ind], tr.Servers[ind+1:]...)
		case models.SubscriptionServerType: //订阅的结点的不能删的
			continue
		}
	}
	err = tr.WriteToFile()
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	config.SetTouchRaw(&tr)
	GetTouch(ctx)
}
