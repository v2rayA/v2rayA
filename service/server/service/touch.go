package service

import (
	"github.com/v2rayA/v2rayA/db/configure"
)

func DeleteWhich(ws []*configure.Which) (err error) {
	var data *configure.Whiches
	// Deduplicate touches to delete
	data = configure.NewWhiches(ws)
	data = configure.NewWhiches(data.GetNonDuplicated())
	// Sort touches to delete, placing larger indices first to delete from back to front
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
		case configure.SubscriptionType: // Here a subscription is being deleted
			// Check if the currently connected node is in this subscription, if so, disconnect
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
			// Check if the currently connected node is this server, if so, disconnect
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
