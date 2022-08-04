package db

import (
	"fmt"
	"go.etcd.io/bbolt"
	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"reflect"
	"sort"
	"strconv"
)

func ListSet(bucket string, key string, index int, val interface{}) (err error) {
	return DB().Update(func(tx *bbolt.Tx) error {
		if bkt, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
			return err
		} else {
			b := bkt.Get([]byte(key))
			if b == nil {
				return fmt.Errorf("ListSet: can't set element to an empty list")
			}

			if b, e := sjson.SetBytes(b, strconv.Itoa(index), val); e != nil {
				return e
			} else {
				return bkt.Put([]byte(key), b)
			}
		}
	})
}

func ListGet(bucket string, key string, index int) (b []byte, err error) {
	err = DB().Update(func(tx *bbolt.Tx) error {
		if bkt, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
			return err
		} else {
			v := bkt.Get([]byte(key))
			if v == nil {
				return fmt.Errorf("ListGet: can't get element from an empty list")
			}
			r := gjson.GetBytes(v, strconv.Itoa(index))
			if r.Exists() {
				b = []byte(r.Raw)
				return nil
			} else {
				return fmt.Errorf("ListGet: no such element")
			}
		}
	})
	return b, err
}

func ListAppend(bucket string, key string, val interface{}) (err error) {
	return DB().Update(func(tx *bbolt.Tx) error {
		if bkt, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
			return err
		} else {
			bList := bkt.Get([]byte(key))
			if bList == nil {
				bList = []byte("[]")
			}
			v := reflect.ValueOf(val)
			if v.Kind() == reflect.Slice {
				sliceLen := v.Len()
				for i := 0; i < sliceLen; i++ {
					if bList, err = sjson.SetBytes(bList, "-1", v.Index(i).Interface()); err != nil {
						return err
					}
				}
			} else {
				if bList, err = sjson.SetBytes(bList, "-1", val); err != nil {
					return err
				}
			}
			return bkt.Put([]byte(key), bList)
		}
	})
}

func ListGetAll(bucket string, key string) (list [][]byte, err error) {
	err = DB().Update(func(tx *bbolt.Tx) error {
		if bkt, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
			return err
		} else {
			b := bkt.Get([]byte(key))
			if b == nil {
				return nil
			}
			parsed := gjson.ParseBytes(b)
			if !parsed.IsArray() {
				return fmt.Errorf("ListGetAll: is not array")
			}
			results := parsed.Array()
			for _, r := range results {
				list = append(list, []byte(r.Raw))
			}
		}
		return nil
	})
	return list, err
}

func ListRemove(bucket, key string, indexes []int) error {
	if len(indexes) == 0 {
		return fmt.Errorf("ListRemove: nothing to remove")
	}
	return DB().Update(func(tx *bbolt.Tx) error {
		if bkt, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
			return err
		} else {
			b := bkt.Get([]byte(key))

			var list []interface{}
			if err := jsoniter.Unmarshal(b, &list); err != nil {
				return err
			}
			sort.Ints(indexes)
			maxIndexToDelete := indexes[len(indexes)-1]
			if maxIndexToDelete >= len(list) || indexes[0] < 0 {
				return fmt.Errorf("ListRemove: index out of range")
			}
			j := 1
			dist := 1
			for i := indexes[0]; i+dist < len(list); i++ {
				for j < len(indexes) && i+dist >= indexes[j] {
					if indexes[j] != indexes[j-1] {
						dist++
					}
					j++
				}
				if i+dist >= len(list) {
					break
				}
				list[i] = list[i+dist]
			}
			list = list[:len(list)-dist]

			b, _ = jsoniter.Marshal(list)
			return bkt.Put([]byte(key), b)
		}
	})
}

func ListLen(bucket string, key string) (length int, err error) {
	err = DB().Update(func(tx *bbolt.Tx) error {
		if bkt, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
			return err
		} else {
			b := bkt.Get([]byte(key))
			if b == nil {
				return nil
			}
			parsed := gjson.ParseBytes(b)
			if !parsed.IsArray() {
				return fmt.Errorf("ListLen: is not array")
			}
			length = len(parsed.Array())
		}
		return nil
	})
	return length, err
}
