package repository_test

import (
	"testing"

	"github.com/stjudewashere/seonaut/internal/repository"
)

const (
	url1        = "https://example.com"
	url2        = "https://example.com/hash"
	longstring  = "abcdeabcdeabcdeabcdeabcdeabcdeabcdeabcdeabcdeabcdeabcdeabcde"
	shortstring = "short"
	hash1       = "100680ad546ce6a577f42f52df33b4cfdca756859e664b8d7de329b150d09ce9"
	hash2       = "73d942d72d2df275546b54948c19f71112007be1bba007a082563a17957cdcaa"
)

// Test the hash function is actually hashing the strings"
func TestHash(t *testing.T) {
	h := repository.Hash(url1)
	if h != hash1 {
		t.Error("Error hashing url1")
	}

	h = repository.Hash(url2)
	if h != hash2 {
		t.Error("Error hashing url2")
	}
}

// Test the truncate function for both truncated and non-truncated text.
func TestTruncate(t *testing.T) {
	truncated := repository.Truncate(longstring, 6)
	if truncated != "abc..." {
		t.Error("Error truncating string")
	}

	truncated = repository.Truncate(shortstring, 500)
	if truncated != shortstring {
		t.Error("Error truncating short string")
	}
}
