package db

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"go.etcd.io/bbolt"
	"github.com/v2rayA/v2rayA/common"
)

type set map[[32]byte]interface{}

func bytes2Sha256(b []byte) [32]byte {
	h := sha256.New()
	h.Write(b)
	var hash [32]byte
	copy(hash[:], h.Sum(nil))
	return hash
}

func toSha256(val interface{}) (hash [32]byte, err error) {
	b, err := common.ToBytes(val)
	if err != nil {
		return hash, err
	}
	hash = bytes2Sha256(b)
	return hash, nil
}

func setOp(bucket string, key string, f func(m set) (readonly bool, err error)) (err error) {
	return DB().Update(func(tx *bbolt.Tx) error {
		if bkt, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
			return err
		} else {
			var m set
			v := bkt.Get([]byte(key))
			if v == nil {
				m = make(set)
			} else if err := gob.NewDecoder(bytes.NewReader(v)).Decode(&m); err != nil {
				return err
			}
			if readonly, err := f(m); err != nil {
				return err
			} else if readonly {
				return nil
			}
			if b, err := common.ToBytes(m); err != nil {
				return err
			} else {
				return bkt.Put([]byte(key), b)
			}
		}
	})
}

func SetAdd(bucket string, key string, val interface{}) (err error) {
	h, err := toSha256(val)
	if err != nil {
		return err
	}
	return setOp(bucket, key, func(m set) (readonly bool, err error) {
		m[h] = val
		return false, nil
	})
}

func SetRemove(bucket string, key string, val interface{}) (err error) {
	h, err := toSha256(val)
	if err != nil {
		return err
	}
	return setOp(bucket, key, func(m set) (readonly bool, err error) {
		if _, ok := m[h]; ok {
			delete(m, h)
		}
		return false, nil
	})
}

func StringSetGetAll(bucket string, key string) (members []string, err error) {
	err = setOp(bucket, key, func(m set) (readonly bool, err error) {
		for _, v := range m {
			members = append(members, v.(string))
		}
		return true, nil
	})
	if err != nil {
		return nil, err
	}
	return members, nil
}
