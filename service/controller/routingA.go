package controller

import (
	"V2RayA/core/routingA"
	"V2RayA/persistence/configure"
	"V2RayA/tools"
	"errors"
	"github.com/gin-gonic/gin"
)

func GetRoutingA(ctx *gin.Context) {
	tools.ResponseSuccess(ctx, gin.H{
		"routingA": configure.GetRoutingA(),
	})
}
func PutRoutingA(ctx *gin.Context) {
	var data struct {
		RoutingA string `json:"routingA"`
	}
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		tools.ResponseError(ctx, errors.New("bad request"+err.Error()))
		return
	}
	_, err = routingA.Parse(data.RoutingA)
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	err = configure.SetRoutingA(data.RoutingA)
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	tools.ResponseSuccess(ctx, nil)
}
