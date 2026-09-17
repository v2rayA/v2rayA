package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"sync"

	"github.com/v2rayA/v2rayA/db"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

var secret []byte
var once sync.Once

const (
	secretBucket = "system"
	secretKey    = "jwtSecret"
)

// genSecret loads the signing secret, generating and storing a random one on
// first use. It used to be the hash of the subscription addresses, or of the
// server hostnames when there was no subscription, so importing or deleting a
// server changed the secret and signed every session out.
func genSecret() {
	var stored string
	if err := db.Get(secretBucket, secretKey, &stored); err == nil && stored != "" {
		if b, err := hex.DecodeString(stored); err == nil && len(b) == 32 {
			secret = b
			return
		}
	}
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		log.Fatal("failed to generate the session secret: %v", err)
	}
	if err := db.Set(secretBucket, secretKey, hex.EncodeToString(b)); err != nil {
		// Without persistence every restart signs sessions out, which is
		// still better than refusing to start.
		log.Warn("failed to store the session secret, sessions will not survive a restart: %v", err)
	}
	secret = b
}

func getSecret() []byte {
	once.Do(genSecret)
	return secret
}
