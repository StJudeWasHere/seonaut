package helper

import (
	"crypto/sha256"
	"encoding/hex"
)

// Hash returns a hashed string
func Hash(s string) string {
	hash := sha256.Sum256([]byte(s))

	return hex.EncodeToString(hash[:])
}
