package handlers

import (
	"V2RayA/config"
	"V2RayA/models"
	"V2RayA/tools"
	"github.com/gin-gonic/gin"
	"sort"
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
			t.Subscriptions[i].Servers[j].TYPE = models.SubscriptionType
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
	// TODO: 特判删除connected节点时的情况
	var data models.WhichTouches
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	tr := config.GetTouchRaw()
	//对要删除的touch去重，并将ID转化为下标，然后检查下标有效性
	ts := make(map[models.WhichTouch]struct{})
	for i := range data.Touches {
		ind := data.Touches[i].ID - 1
		v := data.Touches[i]
		switch v.TYPE {
		case models.SubscriptionType:
			if ind >= 0 && ind < len(tr.Subscriptions) {
				ts[v] = struct{}{}
			}
		case models.ServerType:
			if ind >= 0 && ind < len(tr.Servers) {
				ts[v] = struct{}{}
			}
		}
	}
	data.Touches = make([]models.WhichTouch, 0)
	for k := range ts {
		data.Touches = append(data.Touches, k)
	}
	//对要删除的touch排序，将大的下标排在前面，从后往前删
	sort.Sort(data)
	touches := data.GetTouches()
	tr.Lock() //写操作需要上锁
	defer tr.Unlock()
	for _, v := range touches {
		ind := v.ID - 1
		switch v.TYPE {
		case models.SubscriptionType:
			tr.Subscriptions = append(tr.Subscriptions[:ind], tr.Subscriptions[ind+1:]...)
		case models.ServerType:
			tr.Servers = append(tr.Servers[:ind], tr.Servers[ind+1:]...)
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
