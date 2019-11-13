package router

import (
	"V2RayA/global"
	"V2RayA/handlers"
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
		g.GET("version", handlers.Version)
		g.GET("resolving", handlers.Resolving)
		g.POST("import", handlers.Import)
		g.GET("touch", handlers.GetTouch)
		g.DELETE("touch", handlers.DeleteTouch)
		g.POST("connection", handlers.PostConnection)
		g.DELETE("connection", handlers.DeleteConnection)
		g.POST("v2ray", handlers.PostV2ray)
		g.DELETE("v2ray", handlers.DeleteV2ray)
		g.GET("pingLatency", handlers.GetPingLatency)
		g.GET("sharingAddress", handlers.GetSharingAddress)
		g.GET("remoteGFWListVersion", handlers.GetRemoteGFWListVersion)
		g.GET("setting", handlers.GetSetting)
		g.PUT("setting", handlers.PutSetting)
		g.PUT("gfwList", handlers.PutGFWList)
		g.PUT("subscription", handlers.PutSubscription)
	}
	log.Fatal(engine.Run(fmt.Sprintf("%v:%v", app.Address, app.Port)))
}
