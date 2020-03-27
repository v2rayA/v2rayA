package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"runtime/pprof"
)

func main() {
	f, err := os.Create("/tmp/v2raya.prof")
	if err != nil {
		panic(err)
	}
	err = pprof.StartCPUProfile(f)
	if err != nil {
		panic(err)
	}
	defer pprof.StopCPUProfile()


	gin.SetMode(gin.ReleaseMode)
	checkEnvironment()
	initConfigure()
	checkConnection()
	hello()
	checkUpdate()
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
