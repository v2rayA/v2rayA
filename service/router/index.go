package router

import (
	"V2RayA/controller"
	"V2RayA/global"
	"V2RayA/persistence/configure"
	"V2RayA/tools"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gookit/color"
)

func Run() error {
	engine := gin.Default()
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AddAllowHeaders("Authorization")
	engine.Use(cors.New(corsConfig))
	engine.GET("/", func(ctx *gin.Context) {
		ctx.String(200, `这里是V2RayA服务端，请配合前端GUI使用，方法：https://github.com/mzz2017/V2RayA/blob/master/README.md`)
	})
	noAuth := engine.Group("api")
	{
		noAuth.GET("version", controller.GetVersion)
		noAuth.POST("login", controller.PostLogin)
		noAuth.POST("account", controller.PostAccount)
	}
	auth := engine.Group("api")
	auth.Use(func(ctx *gin.Context) {
		if !configure.HasAnyAccounts() {
			tools.Response(ctx, tools.UNAUTHORIZED, gin.H{
				"first": true,
			})
			ctx.Abort()
			return
		}
		ctx.Next()
	}, tools.JWTAuth(false))
	{
		auth.GET("resolving", controller.GetResolving)
		auth.POST("import", controller.PostImport)
		auth.GET("touch", controller.GetTouch)
		auth.DELETE("touch", controller.DeleteTouch)
		auth.POST("connection", controller.PostConnection)
		auth.DELETE("connection", controller.DeleteConnection)
		auth.POST("v2ray", controller.PostV2ray)
		auth.DELETE("v2ray", controller.DeleteV2ray)
		auth.GET("pingLatency", controller.GetPingLatency)
		auth.GET("sharingAddress", controller.GetSharingAddress)
		auth.GET("remoteGFWListVersion", controller.GetRemoteGFWListVersion)
		auth.GET("setting", controller.GetSetting)
		auth.PUT("setting", controller.PutSetting)
		auth.PUT("gfwList", controller.PutGFWList)
		auth.PUT("subscription", controller.PutSubscription)
		noAuth.GET("ports", controller.GetPorts)
		noAuth.PUT("ports", controller.PutPorts)
		noAuth.PUT("account", controller.PutAccount)
	}
	color.Red.Println("GUI demo: https://v2raya.mzz.pub")
	app := global.GetEnvironmentConfig()
	return engine.Run(app.Address)
}
