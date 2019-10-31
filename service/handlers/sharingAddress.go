package handlers

import (
	"V2RayA/config"
	"V2RayA/models"
	"V2RayA/tools"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
)

func GetSharingAddress(ctx *gin.Context) {
	var wt models.WhichTouch
	err := json.Unmarshal([]byte(ctx.Query("touch")), &wt)
	if err != nil {
		tools.ResponseError(ctx, errors.New("参数有误"))
		return
	}
	tr := config.GetTouchRaw()
	var add string
	if wt.TYPE == models.SubscriptionType {
		ind := wt.ID - 1
		if ind < 0 || ind >= len(tr.Subscriptions) {
			tools.ResponseError(ctx, errors.New("参数有误"))
			return
		}
		add = tr.Subscriptions[ind].Address
	} else {
		tsr, err := tr.LocateServer(&wt)
		if err != nil {
			tools.ResponseError(ctx, err)
			return
		}
		add = tools.GenerateURL(tsr.VmessInfo)
		if add == "" {
			tools.ResponseError(ctx, errors.New("生成地址时发生错误"))
			return
		}
	}
	tools.ResponseSuccess(ctx, gin.H{"sharingAddress": add})
}
