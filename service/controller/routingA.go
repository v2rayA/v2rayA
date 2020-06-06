package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/mzz2017/v2rayA/common"
	"github.com/mzz2017/v2rayA/core/routingA"
	"github.com/mzz2017/v2rayA/db/configure"
)

func GetRoutingA(ctx *gin.Context) {
	common.ResponseSuccess(ctx, gin.H{
		"routingA": configure.GetRoutingA(),
	})
}
func PutRoutingA(ctx *gin.Context) {
	var data struct {
		RoutingA string `json:"routingA"`
	}
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		common.ResponseError(ctx, logError(err, "bad request"))
		return
	}
	_, err = routingA.Parse(data.RoutingA)
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	err = configure.SetRoutingA(&data.RoutingA)
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	common.ResponseSuccess(ctx, nil)
}
