package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"sync"

	"github.com/v2rayA/v2rayA/db"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

const (
	secretBucket = "system"
	secretKey    = "jwtSecret"
)

var (
	secretMu  sync.Mutex
	secret    []byte
	persisted bool
)

// loadSecret returns the signing secret, generating and storing a random one on
// first use. It used to be the hash of the subscription addresses, or of the
// server hostnames when there was no subscription, so importing or deleting a
// server changed the secret and signed every session out.
func loadSecret() []byte {
	secretMu.Lock()
	defer secretMu.Unlock()
	if secret != nil && persisted {
		return secret
	}
	var stored string
	err := db.Get(secretBucket, secretKey, &stored)
	if err == nil {
		if b, e := hex.DecodeString(stored); e == nil && len(b) == 32 {
			secret = b
			persisted = true
			return secret
		}
		log.Warn("the stored session secret is unusable, generating a new one")
	} else if !isNotFound(err) {
		// The row may exist and be unreadable right now. Overwriting it would
		// invalidate every session, so sign with a temporary secret and try
		// again on the next call.
		if secret == nil {
			secret = randomSecret()
		}
		log.Warn("cannot read the session secret, sessions will not survive a restart: %v", err)
		return secret
	}
	if secret == nil {
		secret = randomSecret()
	}
	if err := db.Set(secretBucket, secretKey, hex.EncodeToString(secret)); err != nil {
		log.Warn("failed to store the session secret, sessions will not survive a restart: %v", err)
		return secret
	}
	persisted = true
	return secret
}

func randomSecret() []byte {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		log.Fatal("failed to generate the session secret: %v", err)
	}
	return b
}

func isNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "key is not found")
}

func getSecret() []byte {
	return loadSecret()
}
