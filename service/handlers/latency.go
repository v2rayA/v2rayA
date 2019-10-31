package handlers

import (
	"V2RayA/config"
	"V2RayA/models"
	"V2RayA/tools"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sparrc/go-ping"
	"sync"
	"time"
)

func GetPingLatency(ctx *gin.Context) {
	var data models.WhichTouches
	err := json.Unmarshal([]byte(ctx.Query("touches")), &data.Touches)
	if err != nil {
		tools.ResponseError(ctx, errors.New("参数有误"))
		return
	}
	tr := config.GetTouchRaw()
	//对要Ping的touch去重
	data.SetTouches(data.GetNonDuplicatedTouches(&tr))
	touches := data.GetTouches()
	//多线程异步ping
	wg := new(sync.WaitGroup)
	for i, v := range touches {
		if v.TYPE == models.SubscriptionType { //subscription不能ping
			continue
		}
		tsr, err := tr.LocateServer(&v)
		if err != nil {
			touches[i].PingLatency = new(string)
			*touches[i].PingLatency = "backend server error"
			continue
		}
		pinger, err := ping.NewPinger(tsr.VmessInfo.Add)
		if err != nil {
			tools.ResponseError(ctx, err)
			return
		}
		pinger.Count = 5
		pinger.Timeout = time.Second * 5
		pinger.SetPrivileged(true)
		wg.Add(1)
		go func(pinger *ping.Pinger, i int) {
			pinger.Run()
			s := pinger.Statistics()
			touches[i].PingLatency = new(string)
			*touches[i].PingLatency = fmt.Sprintf("平均: %dms, 最快: %dms, 最慢: %dms. 丢包: %d/%d(%.1f%%)", int(s.AvgRtt.Seconds()*1000), int(s.MaxRtt.Seconds()*1000), int(s.MinRtt.Seconds()*1000), s.PacketsSent-s.PacketsRecv, s.PacketsSent, s.PacketLoss)
			wg.Done()
		}(pinger, i)
	}
	wg.Wait()
	for i := len(data.Touches) - 1; i >= 0; i-- {
		if data.Touches[i].TYPE == models.SubscriptionType { //不返回subscriptionType
			data.Touches = append(data.Touches[:i], data.Touches[i+1:]...)
		}
	}
	tools.ResponseSuccess(ctx, data)
}
