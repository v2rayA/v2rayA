package db

import (
	"go.etcd.io/bbolt"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/pkg/util/copyfile"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"os"
	"path/filepath"
	"sync"
)

var once sync.Once
var db *bbolt.DB
var readOnly bool

func SetReadOnly() {
	readOnly = true
}

func initDB() {
	confPath := conf.GetEnvironmentConfig().Config
	dbPath := filepath.Join(confPath, "boltv4.db")
	if readOnly {
		// trick: not really read-only
		f, err := os.CreateTemp(os.TempDir(), "v2raya_tmp_boltv4_*.db")
		if err != nil {
			panic(err)
		}
		newPath := f.Name()
		f.Close()
		if err = copyfile.CopyFileContent(dbPath, newPath); err != nil {
			panic(err)
		}
		dbPath = newPath
	}

	var err error
	db, err = bbolt.Open(dbPath, 0600, nil)
	if err != nil {
		log.Fatal("bbolt.Open: %v", err)
	}
}

func DB() *bbolt.DB {
	once.Do(initDB)
	return db
}
