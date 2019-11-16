package router

import (
	"V2RayA/global"
	"V2RayA/controller"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
)

func Run() {
	app := global.GetServiceConfig()
	engine := gin.New()
	engine.Use(gin.Recovery())
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	engine.Use(cors.New(corsConfig))
	engine.GET("/", func(ctx *gin.Context) {
		ctx.String(200, `这里是V2RayA服务端，请配合前端GUI使用，方法：https://github.com/mzz2017/V2RayA/blob/master/README.md`)
	})
	g := engine.Group("api")
	{
		g.GET("version", controller.Version)
		g.GET("resolving", controller.Resolving)
		g.POST("import", controller.Import)
		g.GET("touch", controller.GetTouch)
		g.DELETE("touch", controller.DeleteTouch)
		g.POST("connection", controller.PostConnection)
		g.DELETE("connection", controller.DeleteConnection)
		g.POST("v2ray", controller.PostV2ray)
		g.DELETE("v2ray", controller.DeleteV2ray)
		g.GET("pingLatency", controller.GetPingLatency)
		g.GET("sharingAddress", controller.GetSharingAddress)
		g.GET("remoteGFWListVersion", controller.GetRemoteGFWListVersion)
		g.GET("setting", controller.GetSetting)
		g.PUT("setting", controller.PutSetting)
		g.PUT("gfwList", controller.PutGFWList)
		g.PUT("subscription", controller.PutSubscription)
	}
	log.Fatal(engine.Run(fmt.Sprintf("%v:%v", app.Address, app.Port)))
}
