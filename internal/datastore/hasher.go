package datastore

import (
	"crypto/sha256"
	"encoding/hex"
)

func Hash(s string) string {
	hash := sha256.Sum256([]byte(s))

	return hex.EncodeToString(hash[:])
}
