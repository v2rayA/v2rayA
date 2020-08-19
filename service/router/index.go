package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gookit/color"
	"github.com/mzz2017/v2rayA/common"
	"github.com/mzz2017/v2rayA/common/jwt"
	"github.com/mzz2017/v2rayA/common/netTools"
	"github.com/mzz2017/v2rayA/controller"
	"github.com/mzz2017/v2rayA/db/configure"
	"github.com/mzz2017/v2rayA/global"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

func ServeGUI(engine *gin.Engine) {
	webDir := path.Join(global.GetEnvironmentConfig().Config, "web")
	filepath.Walk(webDir, func(path string, info os.FileInfo, err error) error {
		if path == webDir {
			return nil
		}
		if info.IsDir() {
			engine.Static("/"+info.Name(), path)
			return filepath.SkipDir
		}
		engine.StaticFile("/"+info.Name(), path)
		return nil
	})
	engine.LoadHTMLFiles(path.Join(webDir, "index.html"))
	engine.GET("/", func(context *gin.Context) {
		context.HTML(http.StatusOK, "index.html", nil)
	})
}

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
		auth.GET("dnsList", controller.GetDnsList)
		auth.PUT("dnsList", controller.PutDnsList)
		auth.GET("siteDatFiles", controller.GetSiteDatFiles)
		auth.GET("customPac", controller.GetCustomPac)
		auth.PUT("customPac", controller.PutCustomPac)
		auth.GET("routingA", controller.GetRoutingA)
		auth.PUT("routingA", controller.PutRoutingA)
	}

	ServeGUI(engine)

	app := global.GetEnvironmentConfig()

	ip, port := netTools.ParseAddress(app.Address)
	addrs, err := net.InterfaceAddrs()
	if net.ParseIP(ip).IsUnspecified() && err == nil {
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				if ipnet.IP.IsLoopback() {
					printRunningAt("http://" + app.Address)
					continue
				}
				printRunningAt("http://" + ipnet.IP.String() + ":" + port)
			}
		}
	} else {
		printRunningAt("http://" + app.Address)
	}
	return engine.Run(app.Address)
}

func printRunningAt(address string) {
	color.Red.Println("v2rayA is listening at", address)
}
