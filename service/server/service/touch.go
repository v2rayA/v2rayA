package service

import (
	"github.com/v2rayA/v2rayA/db/configure"
)

func DeleteWhich(ws []*configure.Which) (err error) {
	var data *configure.Whiches
	//对要删除的touch去重
	data = configure.NewWhiches(ws)
	data = configure.NewWhiches(data.GetNonDuplicated())
	//对要删除的touch排序，将大的下标排在前面，从后往前删
	data.SortSameTypeReverse()
	touches := data.Get()
	cssRaw := configure.GetConnectedServers()
	cssAfter := cssRaw.Get()
	subscriptionsIndexes := make([]int, 0, len(ws))
	serversIndexes := make([]int, 0, len(ws))
	bDeletedSubscription := false
	bDeletedServer := false
	for _, v := range touches {
		ind := v.ID - 1
		switch v.TYPE {
		case configure.SubscriptionType: //这里删的是某个订阅
			//检查现在连接的结点是否在该订阅中，是的话断开连接
			css := cssRaw.Get()
			for i := len(css) - 1; i >= 0; i-- {
				cs := css[i]
				if cs != nil && cs.TYPE == configure.SubscriptionServerType {
					if ind == cs.Sub {
						err = Disconnect(*cs, false)
						if err != nil {
							return
						}
						cssAfter = append(cssAfter[:i], cssAfter[i+1:]...)
					} else if ind < cs.Sub {
						cs.Sub--
					}
				}
			}
			subscriptionsIndexes = append(subscriptionsIndexes, ind)
			bDeletedSubscription = true
		case configure.ServerType:
			//检查现在连接的结点是否是该服务器，是的话断开连接
			css := cssRaw.Get()
			for i := len(css) - 1; i >= 0; i-- {
				cs := css[i]
				if cs != nil && cs.TYPE == configure.ServerType {
					if v.ID == cs.ID {
						err = Disconnect(*cs, false)
						if err != nil {
							return
						}
						cssAfter = append(cssAfter[:i], cssAfter[i+1:]...)
					} else if v.ID < cs.ID {
						cs.ID--
					}
				}
			}
			serversIndexes = append(serversIndexes, ind)
			bDeletedServer = true
		case configure.SubscriptionServerType:
			continue
		}
	}
	if err := configure.OverwriteConnects(configure.NewWhiches(cssAfter)); err != nil {
		return err
	}
	if bDeletedSubscription {
		err = configure.RemoveSubscriptions(subscriptionsIndexes)
		if err != nil {
			return
		}
	}
	if bDeletedServer {
		err = configure.RemoveServers(serversIndexes)
		if err != nil {
			return
		}
	}
	return
}
