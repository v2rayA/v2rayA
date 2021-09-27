package jwt

import (
	"crypto/sha256"
	gonanoid "github.com/matoous/go-nanoid"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"sync"
)

var secret []byte
var once sync.Once

func genSecret() {
	// In order to reduce the number of times to enter the password,
	// if there is a subscription, the secret is the hash value of all subscription addresses.
	// Otherwise, the hash value of all server addresses.
	if sub := configure.GetSubscriptionsV2(); len(sub) > 0 {
		sha := sha256.New()
		for _, s := range sub {
			sha.Write([]byte(s.Address))
		}
		secret = sha.Sum(nil)
	} else if servers := configure.GetServersV2(); len(servers) > 0 {
		sha := sha256.New()
		for _, s := range servers {
			sha.Write([]byte(s.ServerObj.GetHostname()))
		}
		secret = sha.Sum(nil)
	} else {
		id, err := gonanoid.Nanoid()
		if err != nil {
			log.Fatal("failed to genSecret: %v", err)
		}
		secret = []byte(id)
	}
}
func getSecret() []byte {
	once.Do(genSecret)
	return secret
}
