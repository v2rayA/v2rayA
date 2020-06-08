package db

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/xujiajun/nutsdb"
	"log"
	"reflect"
	"sort"
	"v2ray.com/core/common/errors"
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
	var entry *nutsdb.Entry
	if err := DB().View(func(tx *nutsdb.Tx) error {
		var e error
		if entry, e = tx.Get(bucket, []byte(key)); e != nil {
			return e
		}
		return nil
	}); err != nil {
		if errors.Cause(err) != nutsdb.ErrBucketAndKey(bucket, []byte(key)) {
			log.Println(newError("[ERROR] func Exists returns a new error type").Base(err))
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
			return newError().Base(err)
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
				return newError().Base(err)
			}
		}
		length = len(entries)
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
			return newError().Base(err)
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
			return newError().Base(err)
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
			return newError().Base(err)
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
	// TODO: waiting for https://github.com/xujiajun/nutsdb/issues/66
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
			return newError().Base(err)
		}
		for i := len(indexes) - 1; i >= 0; i-- {
			list = append(list[:indexes[i]], list[indexes[i]+1:]...)
		}
		if err = tx.LRem(bucket, []byte(key), 0); err != nil {
			return newError().Base(err)
		}
		if err = tx.RPush(bucket, []byte(key), list...); err != nil {
			return newError().Base(err)
		}
		return nil
	})
}

func BucketClear(bucket string) error {
	return DB().Update(func(tx *nutsdb.Tx) error {
		entries, err := tx.GetAll(bucket)
		if err != nil {
			return newError().Base(err)
		}
		for _, e := range entries {
			err = tx.Delete(bucket, e.Key)
			if err != nil {
				return newError().Base(err)
			}
		}
		return nil
	})
}
