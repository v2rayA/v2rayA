package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gookit/color"
	"v2rayA/common"
	"v2rayA/common/jwt"
	"v2rayA/controller"
	"v2rayA/global"
	"v2rayA/db/configure"
)

func Run() error {
	engine := gin.New()
	//ginpprof.Wrap(engine)
	engine.Use(gin.Recovery())
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowMethods = []string{
		"GET", "POST", "DELETE", "PUT", "PATCH", "OPTIONS", "HEAD",
	}
	corsConfig.AddAllowHeaders("Authorization")
	engine.Use(cors.New(corsConfig))
	engine.GET("/", func(ctx *gin.Context) {
		ctx.Header("Content-Type", "text/html; charset=utf-8")
		ctx.String(418, `<body>Here is v2rayA backend. Reference: <a href="https://github.com/mzz2017/v2rayA">https://github.com/mzz2017/v2rayA</a></body>`)
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
			common.Response(ctx, common.UNAUTHORIZED, gin.H{
				"first": true,
			})
			ctx.Abort()
			return
		}
		ctx.Next()
	}, jwt.JWTAuth(false))
	{
		//auth.GET("resolving", controller.GetResolving)
		auth.POST("import", controller.PostImport)
		auth.GET("touch", controller.GetTouch)
		auth.DELETE("touch", controller.DeleteTouch)
		auth.POST("connection", controller.PostConnection)
		auth.DELETE("connection", controller.DeleteConnection)
		auth.POST("v2ray", controller.PostV2ray)
		auth.DELETE("v2ray", controller.DeleteV2ray)
		auth.GET("pingLatency", controller.GetPingLatency)
		auth.GET("httpLatency", controller.GetHttpLatency)
		auth.GET("sharingAddress", controller.GetSharingAddress)
		auth.GET("remoteGFWListVersion", controller.GetRemoteGFWListVersion)
		auth.GET("setting", controller.GetSetting)
		auth.PUT("setting", controller.PutSetting)
		auth.PUT("gfwList", controller.PutGFWList)
		auth.PUT("subscription", controller.PutSubscription)
		auth.PATCH("subscription", controller.PatchSubscription)
		auth.GET("ports", controller.GetPorts)
		auth.PUT("ports", controller.PutPorts)
		//auth.PUT("account", controller.PutAccount)
		auth.GET("portWhiteList", controller.GetPortWhiteList)
		auth.PUT("portWhiteList", controller.PutPortWhiteList)
		auth.POST("portWhiteList", controller.PostPortWhiteList)
		auth.GET("dohList", controller.GetDohList)
		auth.PUT("dohList", controller.PutDohList)
		auth.GET("siteDatFiles", controller.GetSiteDatFiles)
		auth.GET("customPac", controller.GetCustomPac)
		auth.PUT("customPac", controller.PutCustomPac)
		auth.GET("routingA", controller.GetRoutingA)
		auth.PUT("routingA", controller.PutRoutingA)
	}
	color.Red.Println("v2rayA is running at", global.GetEnvironmentConfig().Address)
	color.Red.Println("GUI demo: https://v2raya.mzz.pub")
	color.Red.Println("GUI demo: http://v.mzz.pub")
	app := global.GetEnvironmentConfig()
	return engine.Run(app.Address)
}
