package db

import (
	"github.com/boltdb/bolt"
	"github.com/v2rayA/v2rayA/conf"
	"log"
	"path/filepath"
	"sync"
)

var once sync.Once
var db *bolt.DB

func initDB() {
	confPath := conf.GetEnvironmentConfig().Config
	var err error
	db, err = bolt.Open(filepath.Join(confPath, "boltv4.db"), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func DB() *bolt.DB {
	once.Do(initDB)
	return db
}
