package persistence

import (
	"V2RayA/global"
	"bytes"
	"github.com/json-iterator/go"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"io/ioutil"
	"os"
)

func Get(path string, val interface{}) (err error) {
	b, err := ioutil.ReadFile(global.GetEnvironmentConfig().Config)
	if err != nil {
		return
	}
	v := gjson.GetBytes(b, path)
	if !v.Exists() {
		return newError("bad path")
	}
	return jsoniter.Unmarshal([]byte(v.Raw), val)
}

func Exists(path string) bool {
	b, err := ioutil.ReadFile(global.GetEnvironmentConfig().Config)
	if err != nil {
		return false
	}
	v := gjson.GetBytes(b, path)
	return v.Exists()
}
func GetArrayLen(path string) (length int, err error) {
	b, err := ioutil.ReadFile(global.GetEnvironmentConfig().Config)
	if err != nil {
		return
	}
	v := gjson.GetBytes(b, path)
	if !v.Exists() {
		return -1, newError("bad path")
	}
	if !v.IsArray() {
		return -1, newError("not an array")
	}
	return len(v.Array()), nil
}
func GetObjectLen(path string) (length int, err error) {
	b, err := ioutil.ReadFile(global.GetEnvironmentConfig().Config)
	if err != nil {
		return
	}
	v := gjson.GetBytes(b, path)
	if !v.Exists() {
		return -1, newError("bad path")
	}
	if !v.IsObject() {
		return -1, newError("not an object")
	}
	return len(v.Map()), nil
}

func Set(path string, val interface{}) (err error) {
	if path == "" || path == "." { //这种情况sjson不支持，特判用marshal搞定
		b, _ := jsoniter.Marshal(val)
		return ioutil.WriteFile(global.GetEnvironmentConfig().Config, b, os.FileMode(0600))
	}
	f, err := os.OpenFile(global.GetEnvironmentConfig().Config, os.O_RDWR|os.O_CREATE, os.FileMode(0600))
	if err != nil {
		return
	}
	defer f.Close()
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(f)
	if err != nil {
		return
	}
	b := buf.Bytes()
	b, err = sjson.SetBytes(b, path, val)
	if err != nil {
		return
	}
	err = f.Truncate(0)
	if err != nil {
		return
	}
	_, err = f.WriteAt(b, 0)
	return
}

func Append(path string, val interface{}) (err error) {
	if path == "" || path == "." {
		return newError("cannot append an element at root")
	}
	return Set(path+".-1", val)
}
