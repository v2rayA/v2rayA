package main

import (
	"sync"
	"time"

	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/v2ray/asset/dat"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"github.com/v2rayA/v2rayA/server/service"
)

func updateSubscriptions() {
	subs := configure.GetSubscriptions()
	lenSubs := len(subs)
	control := make(chan struct{}, 2) // concurrency limit: update 2 subscriptions at a time
	// Disconnect from subscriptions before auto-selecting servers from them
	// to limit the number of connected servers and avoid hitting the limit
	shouldDisconnect := true
	err := service.AutoSelectServersFromSubscriptions(shouldDisconnect)
	if err != nil {
		log.Error("[AutoSelect] Failed to disconnect servers from subscriptions -- err: %v", err)
	}
	wg := new(sync.WaitGroup)
	for i := 0; i < lenSubs; i++ {
		wg.Add(1)
		go func(i int) {
			control <- struct{}{}
			err := service.UpdateSubscription(i, false)
			if err != nil {
				log.Info("[AutoUpdate] Subscriptions: Failed to update subscription -- ID: %d, err: %v", i, err)
			} else {
				log.Info("[AutoUpdate] Subscriptions: Complete updating subscription -- ID: %d, Address: %s", i, subs[i].Address)
			}
			wg.Done()
			<-control
		}(i)
	}
	wg.Wait()
	shouldDisconnect = false
	err2 := service.AutoSelectServersFromSubscriptions(shouldDisconnect)
	if err2 != nil {
		log.Error("[AutoSelect] Failed to auto-select servers from subscriptions -- err: %v", err2)
	}

}

func initUpdatingTicker() {
	conf.TickerUpdateGFWList = time.NewTicker(24 * time.Hour * 365 * 100)
	conf.TickerUpdateSubscription = time.NewTicker(24 * time.Hour * 365 * 100)
	go func() {
		for range conf.TickerUpdateGFWList.C {
			_, err := dat.CheckAndUpdateGFWList("")
			if err != nil {
				log.Info("[AutoUpdate] GFWList: %v", err)
			}
		}
	}()
	go func() {
		for range conf.TickerUpdateSubscription.C {
			updateSubscriptions()
		}
	}()
}

func checkUpdate() {
	setting := service.GetSetting()

	// initialize ticker
	initUpdatingTicker()

	// check for PAC file updates
	if setting.GFWListAutoUpdateMode == configure.AutoUpdate ||
		setting.GFWListAutoUpdateMode == configure.AutoUpdateAtIntervals ||
		setting.Transparent == configure.TransparentGfwlist {
		if setting.GFWListAutoUpdateMode == configure.AutoUpdateAtIntervals {
			conf.TickerUpdateGFWList.Reset(time.Duration(setting.GFWListAutoUpdateIntervalHour) * time.Hour)
		}
		switch setting.RulePortMode {
		case configure.GfwlistMode:
			go func() {
				/* Update LoyalsoldierSite.dat */
				localGFWListVersion, err := dat.CheckAndUpdateGFWList("")
				if err != nil {
					log.Warn("Failed to update PAC file: %v", err.Error())
					return
				}
				log.Info("Complete updating PAC file. Localtime: %v", localGFWListVersion)
			}()
		case configure.CustomMode:
			// obsolete
		}
	}

	// check for subscription updates
	if setting.SubscriptionAutoUpdateMode == configure.AutoUpdate ||
		setting.SubscriptionAutoUpdateMode == configure.AutoUpdateAtIntervals {

		if setting.SubscriptionAutoUpdateMode == configure.AutoUpdateAtIntervals {
			conf.TickerUpdateSubscription.Reset(time.Duration(setting.SubscriptionAutoUpdateIntervalHour) * time.Hour)
		}
		go updateSubscriptions()
	}
	// check for server updates
	go func() {
		f := func() {
			if foundNew, remote, err := service.CheckUpdate(); err == nil {
				conf.FoundNew = foundNew
				conf.RemoteVersion = remote
			}
		}
		f()
		c := time.Tick(7 * 24 * time.Hour)
		for range c {
			f()
		}
	}()
}
