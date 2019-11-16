package persistence

import (
	"V2RayA/global"
	"github.com/gomodule/redigo/redis"
	"log"
	"sync"
	"time"
)

var once sync.Once
var pool *redis.Pool
var nextSave = time.Now().Add(30 * 24 * time.Hour)

func saveLater() {
	for {
		if now := time.Now(); now.After(nextSave) {
			nextSave = now.Add(30 * 24 * time.Hour) //设置一个很远的时间
			c := RedisPool().Get()
			_, _ = c.Do("BGSAVE")
		}
		time.Sleep(5 * time.Second)
	}
}

func initRedis() *redis.Pool {
	conf := global.GetServiceConfig()
	p := &redis.Pool{
		Dial: func() (i redis.Conn, e error) {
			return redis.Dial("tcp", conf.RedisServer)
		},
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
	}
	if err := p.Get().Err(); err != nil {
		log.Fatal("redis连接失败: ", err)
	}
	log.Println("redisServer: ", conf.RedisServer)
	go saveLater() //启个协程管理save
	return p
}

func RedisPool() *redis.Pool {
	once.Do(func() {
		pool = initRedis()
	})
	return pool
}

func Do(command string, args ...interface{}) (reply interface{}, err error) {
	c := RedisPool().Get()
	defer c.Close()
	reply, err = c.Do(command, args...)
	return
}

func DoAndSave(command string, args ...interface{}) (reply interface{}, err error) {
	nextSave = time.Now().Add(5 * time.Second) //延迟5秒save，这样如果连续调用本函数，将会推迟save时间
	return Do(command, args...)
}
