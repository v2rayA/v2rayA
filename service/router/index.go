package router

import (
	"V2RayA/config"
	"V2RayA/handlers"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
)

func Run() {
	app := config.GetServiceConfig()
	engine := gin.New()
	engine.Use(gin.Recovery())
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	engine.Use(cors.New(corsConfig))
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
	}
	log.Fatal(engine.Run(fmt.Sprintf("%v:%v", app.Address, app.Port)))
}
