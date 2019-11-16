package service

import (
	"V2RayA/persistence/configure"
)

func DeleteWhich(ws []configure.Which) (err error) {
	var data configure.Whiches
	//对要删除的touch去重
	data.Set(ws)
	data.Set(data.GetNonDuplicated())
	//对要删除的touch排序，将大的下标排在前面，从后往前删
	data.Sort()
	touches := data.Get()
	cs := configure.GetConnectedServer()
	for _, v := range touches {
		ind := v.ID - 1
		switch v.TYPE {
		case configure.SubscriptionType: //这里删的是某个订阅
			//检查现在连接的结点是否在该订阅中，是的话断开连接
			if cs != nil && cs.TYPE == configure.SubscriptionServerType && cs.Sub == ind {
				err = Disconnect()
				if err != nil {
					return
				}
			}
			err = configure.RemoveSubscription(ind)
			if err != nil {
				return
			}
		case configure.ServerType:
			err = configure.RemoveServer(ind)
			if err != nil {
				return
			}
		case configure.SubscriptionServerType: //订阅的结点的不能删的
			continue
		}
	}
	return
}
