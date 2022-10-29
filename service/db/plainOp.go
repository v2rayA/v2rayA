package db

import (
	"fmt"

	jsoniter "github.com/json-iterator/go"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"go.etcd.io/bbolt"
)

func Get(bucket string, key string, val interface{}) (err error) {
	return Transaction(DB(), func(tx *bbolt.Tx) (bool, error) {
		dirty := false
		if bkt, err := CreateBucketIfNotExists(tx, []byte(bucket), &dirty); err != nil {
			return dirty, err
		} else {
			if v := bkt.Get([]byte(key)); v == nil {
				return dirty, fmt.Errorf("Get: key is not found")
			} else {
				return dirty, jsoniter.Unmarshal(v, val)
			}
		}
	})
}

func GetRaw(bucket string, key string) (b []byte, err error) {
	err = Transaction(DB(), func(tx *bbolt.Tx) (bool, error) {
		dirty := false
		if bkt, err := CreateBucketIfNotExists(tx, []byte(bucket), &dirty); err != nil {
			return dirty, err
		} else {
			v := bkt.Get([]byte(key))
			if v == nil {
				return dirty, fmt.Errorf("GetRaw: key is not found")
			}
			b = common.BytesCopy(v)
			return dirty, nil
		}
	})
	return b, err
}

func Exists(bucket string, key string) (exists bool) {
	if err := Transaction(DB(), func(tx *bbolt.Tx) (bool, error) {
		dirty := false
		if bkt, err := CreateBucketIfNotExists(tx, []byte(bucket), &dirty); err != nil {
			return dirty, err
		} else {
			v := bkt.Get([]byte(key))
			exists = v != nil
			return dirty, nil
		}
	}); err != nil {
		log.Warn("%v", err)
		return false
	}
	return exists
}

func GetBucketLen(bucket string) (length int, err error) {
	err = Transaction(DB(), func(tx *bbolt.Tx) (bool, error) {
		dirty := false
		if bkt, err := CreateBucketIfNotExists(tx, []byte(bucket), &dirty); err != nil {
			return dirty, err
		} else {
			length = bkt.Stats().KeyN
		}
		return dirty, nil
	})
	return length, err
}

func GetBucketKeys(bucket string) (keys []string, err error) {
	err = Transaction(DB(), func(tx *bbolt.Tx) (bool, error) {
		dirty := false
		if bkt, err := CreateBucketIfNotExists(tx, []byte(bucket), &dirty); err != nil {
			return dirty, err
		} else {
			return dirty, bkt.ForEach(func(k, v []byte) error {
				keys = append(keys, string(k))
				return nil
			})
		}
	})
	return keys, err
}

func Set(bucket string, key string, val interface{}) (err error) {
	b, err := jsoniter.Marshal(val)
	if err != nil {
		return
	}
	return DB().Update(func(tx *bbolt.Tx) error {
		if bkt, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
			return err
		} else {
			return bkt.Put([]byte(key), b)
		}
	})
}

func BucketClear(bucket string) error {
	err := DB().Update(func(tx *bbolt.Tx) error {
		return tx.DeleteBucket([]byte(bucket))
	})
	if err == bbolt.ErrBucketNotFound {
		return nil
	}
	return err
}
