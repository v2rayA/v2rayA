package service

import (
	"V2RayA/persistence/configure"
	"log"
	"sync"
	"time"
)

func Ping(which []configure.Which, count int, timeout time.Duration) ([]configure.Which, error) {
	var whiches configure.Whiches
	whiches.Set(which)
	//对要Ping的which去重
	which = whiches.GetNonDuplicated()
	//暂时关闭透明代理
	_ = CheckAndStopTransparentProxy()
	defer CheckAndSetupTransparentProxy(true)
	//多线程异步ping
	wg := new(sync.WaitGroup)
	var err error
	for i, v := range which {
		if v.TYPE == configure.SubscriptionType { //subscription不能ping
			continue
		}
		wg.Add(1)
		go func(i int) {
			e := which[i].Ping(count, timeout)
			if e != nil {
				err = e
				log.Println(err)
				//不在乎并发会导致的问题，无需加锁
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	if err != nil {
		return nil, err
	}
	for i := len(which) - 1; i >= 0; i-- {
		if which[i].TYPE == configure.SubscriptionType { //不返回subscriptionType
			which = append(which[:i], which[i+1:]...)
		}
	}
	return which, nil
}
