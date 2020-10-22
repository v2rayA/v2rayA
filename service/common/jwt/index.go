package jwt

import (
	"crypto/sha256"
	"github.com/matoous/go-nanoid"
	"github.com/v2rayA/v2rayA/db/configure"
	"log"
)

var secret []byte

func init() {
	//屡次启动的secret都不一样
	//为了减少输入密码的次数，如果有订阅，secret则为所有订阅地址的hash值
	if sub := configure.GetSubscriptions(); len(sub) > 0 {
		sha := sha256.New()
		for _, s := range sub {
			sha.Write([]byte(s.Address))
		}
		secret = sha.Sum(nil)
	} else {
		id, err := gonanoid.Nanoid()
		if err != nil {
			log.Fatal(err)
		}
		secret = []byte(id)
	}
}
