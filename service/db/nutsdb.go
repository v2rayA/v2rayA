package db

import (
	"github.com/mzz2017/v2rayA/global"
	"github.com/xujiajun/nutsdb"
	"log"
	"os"
	"sync"
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
		_ = os.Chmod(confPath, os.ModeDir|0755)
		log.Fatal(err)
	}
}

func DB() *nutsdb.DB {
	once.Do(initDB)
	return db
}
