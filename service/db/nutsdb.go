package db

import (
	"github.com/xujiajun/nutsdb"
	"log"
	"sync"
	"v2rayA/global"
)

var once sync.Once
var db *nutsdb.DB

func initDB() {
	confPath := global.GetEnvironmentConfig().Config
	var err error
	opt := nutsdb.DefaultOptions
	opt.Dir = confPath
	db, err = nutsdb.Open(opt)
	if err != nil {
		log.Fatal(err)
	}
}

func DB() *nutsdb.DB {
	once.Do(initDB)
	return db
}
