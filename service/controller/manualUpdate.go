package controller

import (
	"V2RayA/persistence/configure"
	"V2RayA/service"
	"V2RayA/tools"
	"errors"
	"github.com/gin-gonic/gin"
)

func PutGFWList(ctx *gin.Context) {
	update, tRemote, err := service.IsUpdate()
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	if update {
		tools.ResponseError(ctx, errors.New(
			"目前最新版本为"+tRemote.Format("2006-01-02")+"，您的本地文件已最新，无需更新",
		))
		return
	}

	/* 更新h2y.dat */
	localGFWListVersion, err := service.UpdateLocalGFWList()
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	tools.ResponseSuccess(ctx, gin.H{
		"localGFWListVersion": localGFWListVersion,
	})
}

func PutSubscription(ctx *gin.Context) {
	var data configure.Which
	err := ctx.ShouldBindJSON(&data)
	index := data.ID - 1
	if err != nil || data.TYPE != configure.SubscriptionType || index < 0 || index >= configure.GetLenSubscriptions() {
		tools.ResponseError(ctx, errors.New("参数有误"))
		return
	}
	err = service.UpdateSubscription(index)
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	GetTouch(ctx)
}
