package handlers

import (
	"V2RayA/global"
	"V2RayA/models/nodeData"
	"V2RayA/models/touch"
	"V2RayA/tools"
	"errors"
	"github.com/gin-gonic/gin"
	"strings"
)

func Import(ctx *gin.Context) {
	var (
		data struct {
			URL string `json:"url"`
		}
		n   *nodeData.NodeData
		err error
	)
	err = ctx.ShouldBindJSON(&data)
	if err != nil {
		tools.ResponseError(ctx, errors.New("参数有误"))
		return
	}
	tr := global.GetTouchRaw()
	tr.Lock() //写操作需要上锁
	defer tr.Unlock()
	if strings.HasPrefix(data.URL, "vmess://") || strings.HasPrefix(data.URL, "ss://") {
		n, err = tools.ResolveURL(data.URL)
		if err != nil {
			tools.ResponseError(ctx, err)
			return
		}
		//后端NodeData转前端TouchServerRaw压入TouchRaw.Servers
		tr.Servers = append(tr.Servers, n.ToTouchServerRaw())
	} else {
		//不是ss://也不是vmess://，有可能是订阅地址
		if !strings.HasPrefix(data.URL, "http://") && !strings.HasPrefix(data.URL, "https://") {
			data.URL = "http://" + data.URL
		}
		c, err := tools.GetHttpClientAutomatically()
		if err != nil {
			tools.ResponseError(ctx, errors.New("尝试使用代理失败，建议修改设置为直连模式再试"))
			return
		}
		infos, err := tools.ResolveSubscriptionWithClient(data.URL, c)
		if err != nil {
			tools.ResponseError(ctx, errors.New("解析订阅地址失败"))
			return
		}
		//后端NodeData转前端TouchServerRaw压入TouchRaw.Subscriptions.Servers
		servers := make([]touch.TouchServerRaw, len(infos))
		for i, v := range infos {
			servers[i] = v.ToTouchServerRaw()
		}
		tr.Subscriptions = append(tr.Subscriptions, touch.SubscriptionRaw{
			Address: data.URL,
			Status:  touch.NewUpdateStatus(),
			Servers: servers,
		})
	}
	//保存到文件
	err = tr.WriteToFile()
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	//录入成功，直接调用Touch接口返回更新后的数据
	global.SetTouchRaw(&tr)
	GetTouch(ctx)
}
