package configure

import (
	"fmt"
	"strconv"

	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/v2rayA/v2rayA/db"
	"go.etcd.io/bbolt"
)

// SetSubscriptionAndConnects keeps server indices and their references in one transaction.
func SetSubscriptionAndConnects(index int, sub *SubscriptionRaw, connects *Whiches) error {
	byOutbound := make(map[string]*Whiches)
	for _, out := range GetOutbounds() {
		byOutbound[out] = NewWhiches(nil)
	}
	for _, w := range connects.Get() {
		if byOutbound[w.Outbound] == nil {
			byOutbound[w.Outbound] = NewWhiches(nil)
		}
		byOutbound[w.Outbound].Add(*w)
	}
	return db.DB().Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("touch"))
		if bucket == nil {
			return fmt.Errorf("subscriptions not initialized")
		}
		list := bucket.Get([]byte("subscriptions"))
		key := strconv.Itoa(index)
		if index < 0 || !gjson.GetBytes(list, key).Exists() {
			return fmt.Errorf("subscription index out of range")
		}
		updated, err := sjson.SetBytes(list, key, sub)
		if err != nil {
			return err
		}
		if err = bucket.Put([]byte("subscriptions"), updated); err != nil {
			return err
		}
		for out, ws := range byOutbound {
			b, err := tx.CreateBucketIfNotExists([]byte("outbound." + out))
			if err != nil {
				return err
			}
			data, err := jsoniter.Marshal(ws)
			if err != nil {
				return err
			}
			if err = b.Put([]byte("connectedServers"), data); err != nil {
				return err
			}
		}
		return nil
	})
}
