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

const subscriptionUpdateCheckInterval = time.Minute

var subscriptionUpdateMu sync.Mutex

func subscriptionShouldUpdate(sub configure.SubscriptionRaw, now time.Time, serviceStart bool) bool {
	switch sub.AutoUpdateMode {
	case configure.AutoUpdate:
		return serviceStart
	case configure.AutoUpdateAtIntervals:
		if sub.AutoUpdateIntervalHour < 1 || sub.LastUpdateAttemptAt == "" {
			return sub.AutoUpdateIntervalHour >= 1
		}
		lastAttempt, err := time.Parse(time.RFC3339, sub.LastUpdateAttemptAt)
		if err != nil {
			return true
		}
		return !now.Before(lastAttempt.Add(time.Duration(sub.AutoUpdateIntervalHour) * time.Hour))
	default:
		return false
	}
}

func updateSubscriptions(serviceStart bool) {
	subscriptionUpdateMu.Lock()
	defer subscriptionUpdateMu.Unlock()

	subs := configure.GetSubscriptions()
	databaseIDs := make([]int64, 0, len(subs))
	now := time.Now()
	for _, sub := range subs {
		if subscriptionShouldUpdate(sub, now, serviceStart) {
			databaseIDs = append(databaseIDs, sub.DatabaseID)
		}
	}
	if len(databaseIDs) == 0 {
		return
	}

	control := make(chan struct{}, 2) // concurrency limit: update 2 subscriptions at a time
	// Disconnect from subscriptions before auto-selecting servers from them
	// to limit the number of connected servers and avoid hitting the limit
	shouldDisconnect := true
	err := service.AutoSelectServersFromSubscriptionIDs(databaseIDs, shouldDisconnect)
	if err != nil {
		log.Error("[AutoSelect] Failed to disconnect servers from subscriptions -- err: %v", err)
	}
	wg := new(sync.WaitGroup)
	for _, databaseID := range databaseIDs {
		wg.Add(1)
		go func(databaseID int64) {
			control <- struct{}{}
			err := service.UpdateSubscriptionByID(databaseID, false)
			if err != nil {
				log.Info("[AutoUpdate] Subscriptions: Failed to update subscription -- ID: %d, err: %v", databaseID, err)
			} else {
				log.Info("[AutoUpdate] Subscriptions: Complete updating subscription -- ID: %d", databaseID)
			}
			wg.Done()
			<-control
		}(databaseID)
	}
	wg.Wait()
	shouldDisconnect = false
	err2 := service.AutoSelectServersFromSubscriptionIDs(databaseIDs, shouldDisconnect)
	if err2 != nil {
		log.Error("[AutoSelect] Failed to auto-select servers from subscriptions -- err: %v", err2)
	}

}

func initUpdatingTicker() {
	conf.TickerUpdateGFWList = time.NewTicker(24 * time.Hour * 365 * 100)
	conf.TickerUpdateSubscription = time.NewTicker(subscriptionUpdateCheckInterval)
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
			updateSubscriptions(false)
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

	go updateSubscriptions(true)
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
