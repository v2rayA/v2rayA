package db

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/pkg/util/copyfile"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"go.etcd.io/bbolt"
)

var once sync.Once
var db *bbolt.DB
var readOnly bool

func SetReadOnly() {
	readOnly = true
}

func initDB() {
	confPath := conf.GetEnvironmentConfig().Config
	dbPath := filepath.Join(confPath, "bolt.db")
	if readOnly {
		// trick: not really read-only
		f, err := os.CreateTemp(os.TempDir(), "v2raya_tmp_bolt_*.db")
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

// The function should return a dirty flag.
// If the dirty flag is true and there is no error then the transaction is commited.
// Otherwise, the transaction is rolled back.
func Transaction(db *bbolt.DB, fn func(*bbolt.Tx) (bool, error)) error {
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	dirty, err := fn(tx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if !dirty {
		return tx.Rollback()
	}
	return tx.Commit()
}

// If the bucket does not exist, the dirty flag is setted
func CreateBucketIfNotExists(tx *bbolt.Tx, name []byte, dirty *bool) (*bbolt.Bucket, error) {
	bkt := tx.Bucket(name)
	if bkt != nil {
		return bkt, nil
	}
	bkt, err := tx.CreateBucket(name)
	if err != nil {
		return nil, err
	}
	*dirty = true
	return bkt, nil
}
