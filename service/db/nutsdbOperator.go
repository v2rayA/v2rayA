package db

import (
	"bytes"
	"encoding/gob"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"github.com/xujiajun/nutsdb"
	"reflect"
	"sort"
)

func Get(bucket string, key string, val interface{}) (err error) {
	var entry *nutsdb.Entry
	if err = DB().View(func(tx *nutsdb.Tx) error {
		var e error
		if entry, e = tx.Get(bucket, []byte(key)); e != nil {
			return e
		}
		return nil
	}); err != nil {
		return
	}
	return jsoniter.Unmarshal(entry.Value, val)
}

func GetRaw(bucket string, key string) (b []byte, err error) {
	err = DB().View(func(tx *nutsdb.Tx) error {
		if entry, e := tx.Get(bucket, []byte(key)); e != nil {
			return e
		} else {
			b = entry.Value
		}
		return nil
	})
	return
}

func Exists(bucket string, key string) bool {
	if err := DB().View(func(tx *nutsdb.Tx) error {
		var e error
		if _, e = tx.Get(bucket, []byte(key)); e != nil {
			return e
		}
		return nil
	}); err != nil {
		if err.Error() != nutsdb.ErrBucketAndKey(bucket, []byte(key)).Error() {
			log.Warn("func Exists returns a new error type: %v", err)
		}
		return false
	}
	return true
}

func ListLen(bucket string, key string) (length int, err error) {
	_ = DB().View(func(tx *nutsdb.Tx) error {
		length, err = tx.LSize(bucket, []byte(key))
		if err == nil {
			return nil
		} else {
			return fmt.Errorf("ListLen: %v", err)
		}
	})
	return
}

func GetBucketLen(bucket string) (length int, err error) {
	_ = DB().View(func(tx *nutsdb.Tx) error {
		var entries nutsdb.Entries
		entries, err = tx.GetAll(bucket)
		if err != nil {
			if err == nil {
				return nil
			} else {
				return fmt.Errorf("GetBucketLen: %v", err)
			}
		}
		length = len(entries)
		return nil
	})
	return
}

func GetBucketKeys(bucket string) (keys []string, err error) {
	_ = DB().View(func(tx *nutsdb.Tx) error {
		var entries nutsdb.Entries
		entries, err = tx.GetAll(bucket)
		if err != nil {
			if err == nil {
				return nil
			} else {
				return fmt.Errorf("GetBucketKeys: %v", err)
			}
		}
		for _, e := range entries {
			keys = append(keys, string(e.Key))
		}
		return nil
	})
	return
}

func Set(bucket string, key string, val interface{}) (err error) {
	b, err := jsoniter.Marshal(val)
	if err != nil {
		return
	}
	return DB().Update(func(tx *nutsdb.Tx) error {
		return tx.Put(bucket, []byte(key), b, nutsdb.Persistent)
	})
}

func SetAdd(bucket string, key string, val interface{}) (err error) {
	buf := new(bytes.Buffer)
	if err = gob.NewEncoder(buf).Encode(val); err != nil {
		return
	}
	return DB().Update(func(tx *nutsdb.Tx) error {
		return tx.SAdd(bucket, []byte(key), buf.Bytes())
	})
}

func SetRemove(bucket string, key string, val interface{}) (err error) {
	buf := new(bytes.Buffer)
	if err = gob.NewEncoder(buf).Encode(val); err != nil {
		return
	}
	return DB().Update(func(tx *nutsdb.Tx) error {
		return tx.SRem(bucket, []byte(key), buf.Bytes())
	})
}

func StringSetGetAll(bucket string, key string) (members []string, err error) {
	err = DB().View(func(tx *nutsdb.Tx) error {
		mbs, err := tx.SMembers(bucket, []byte(key))
		buf := new(bytes.Buffer)
		members = make([]string, len(mbs))
		for i, m := range mbs {
			buf.Reset()
			buf.Write(m)
			if err := gob.NewDecoder(buf).Decode(&members[i]); err != nil {
				return err
			}
		}
		return err
	})
	return
}

func ListSet(bucket string, key string, index int, val interface{}) (err error) {
	b, err := jsoniter.Marshal(val)
	if err != nil {
		return
	}
	return DB().Update(func(tx *nutsdb.Tx) error {
		return tx.LSet(bucket, []byte(key), index, b)
	})
}

func ListGet(bucket string, key string, index int, val interface{}) (err error) {
	var list [][]byte
	return DB().View(func(tx *nutsdb.Tx) error {
		list, err = tx.LRange(bucket, []byte(key), index, index)
		if err == nil {
			err = jsoniter.Unmarshal(list[0], &val)
		}
		if err == nil {
			return nil
		} else {
			return fmt.Errorf("ListGet: %v", err)
		}
	})
}

func ListGetRaw(bucket string, key string, index int) (raw []byte, err error) {
	var list [][]byte
	err = DB().View(func(tx *nutsdb.Tx) error {
		list, err = tx.LRange(bucket, []byte(key), index, index)
		if err == nil {
			raw = list[0]
		}
		if err == nil {
			return nil
		} else {
			return fmt.Errorf("ListGetRaw: %v", err)
		}
	})
	return
}

func ListAppend(bucket string, key string, val interface{}) (err error) {
	var bArray [][]byte
	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Slice {
		sliceLen := v.Len()
		for i := 0; i < sliceLen; i++ {
			b, e := jsoniter.Marshal(v.Index(i).Interface())
			if e != nil {
				return e
			}
			bArray = append(bArray, b)
		}
	} else {
		b, e := jsoniter.Marshal(val)
		if e != nil {
			return e
		}
		bArray = append(bArray, b)
	}
	return DB().Update(func(tx *nutsdb.Tx) error {
		return tx.RPush(bucket, []byte(key), bArray...)
	})
}

func ListGetAll(bucket string, key string) (list [][]byte, err error) {
	err = DB().View(func(tx *nutsdb.Tx) error {
		list, err = tx.LRange(bucket, []byte(key), 0, -1)
		if err == nil {
			return nil
		} else {
			return fmt.Errorf("ListGetAll: %w", err)
		}
	})
	return
}

type beginTo struct {
	begin int
	to    int
}

func indexesToBeginTos(indexes []int) []beginTo {
	sort.Ints(indexes)
	length := len(indexes)
	begin := indexes[0]
	beginTos := make([]beginTo, 0, length/2)
	for i := 1; i < length; i++ {
		if indexes[i]-indexes[i-1] > 1 {
			beginTos = append(beginTos, beginTo{
				begin: begin,
				to:    indexes[i-1],
			})
			begin = indexes[i]
		}
	}
	beginTos = append(beginTos, beginTo{
		begin: begin,
		to:    indexes[length-1],
	})
	return beginTos
}

func ListRemove(bucket, key string, indexes []int) error {
	// TODO: waiting for https://github.com/xujiajun/nutsdb/issues/93
	sort.Ints(indexes)
	return DB().Update(func(tx *nutsdb.Tx) (err error) {
		//for i := len(indexes) - 1; i >= 0; i-- {
		//	err = tx.LSet(bucket, []byte(key), indexes[i], []byte{0})
		//	if err != nil {
		//		return newError().Base(err)
		//	}
		//}
		//tx.LRem(bucket, []byte(key), 0)
		list, err := tx.LRange(bucket, []byte(key), 0, -1)
		if err != nil {
			return fmt.Errorf("ListRemove: %v", err)
		}
		for i := len(indexes) - 1; i >= 0; i-- {
			list = append(list[:indexes[i]], list[indexes[i]+1:]...)
		}
		if err = tx.LRem(bucket, []byte(key), 0); err != nil {
			return fmt.Errorf("ListRemove: %v", err)
		}
		if err = tx.RPush(bucket, []byte(key), list...); err != nil {
			return fmt.Errorf("ListRemove: %v", err)
		}
		return nil
	})
}

func BucketClear(bucket string) error {
	return DB().Update(func(tx *nutsdb.Tx) error {
		entries, err := tx.GetAll(bucket)
		if err != nil {
			if err == nutsdb.ErrBucketEmpty {
				return nil
			}
			return fmt.Errorf("BucketClear: %v", err)
		}
		for _, e := range entries {
			err = tx.Delete(bucket, e.Key)
			if err != nil {
				return fmt.Errorf("BucketClear: %v", err)
			}
		}
		return nil
	})
}
