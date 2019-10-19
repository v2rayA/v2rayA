package router

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"v2rayW/config"
	"v2rayW/handlers"
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
	}
	log.Fatal(engine.Run(fmt.Sprintf("%v:%v", app.Address, app.Port)))
}
