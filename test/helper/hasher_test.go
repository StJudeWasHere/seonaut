package helper

import (
	"testing"

	"github.com/stjudewashere/seonaut/internal/helper"
)

const (
	url1  = "https://example.com"
	url2  = "https://example.com/hash"
	hash1 = "100680ad546ce6a577f42f52df33b4cfdca756859e664b8d7de329b150d09ce9"
	hash2 = "73d942d72d2df275546b54948c19f71112007be1bba007a082563a17957cdcaa"
)

func TestHash(t *testing.T) {
	h := helper.Hash(url1)
	if h != hash1 {
		t.Error("Error hashing url1")
	}

	h = helper.Hash(url2)
	if h != hash2 {
		t.Error("Error hashing url2")
	}
}
