package global

import (
	"V2RayA/models/touch"
	"github.com/mohae/deepcopy"
)

var tr *touch.TouchRaw

/*返回的是一份深拷贝*/
func GetTouchRaw() touch.TouchRaw {
	if tr == nil {
		tr = new(touch.TouchRaw)
		tr.Lock()
		defer tr.Unlock()
		_ = tr.ReadFromFile(GetServiceConfig().ConfigPath)
		if tr.Subscriptions == nil {
			tr.Subscriptions = make([]touch.SubscriptionRaw, 0)
		}
		if tr.Servers == nil {
			tr.Servers = make([]touch.TouchServerRaw, 0)
		}
		if tr.Setting == nil {
			tr.Setting = touch.NewSetting()
		}
	}
	return *deepcopy.Copy(tr).(*touch.TouchRaw)
}

/*更新config中的tr备份*/
func SetTouchRaw(newTr *touch.TouchRaw) {
	tr = newTr
}
