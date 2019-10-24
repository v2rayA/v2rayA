package main

import (
	"V2RayA/router"
	"fmt"
	"os"
)

func main() {
	wd, _ := os.Getwd()
	fmt.Println("working directory is", wd)
	router.Run()
}
