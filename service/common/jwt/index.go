package jwt

import (
	"github.com/matoous/go-nanoid"
	"log"
)

var secret string

func init() {
	//屡次启动的secret都不一样
	id, err := gonanoid.Nanoid()
	if err != nil {
		log.Fatal(err)
	}
	secret = id
}
