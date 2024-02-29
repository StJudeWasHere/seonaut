package models

import (
	"net/url"
)

type Link struct {
	URL        string
	ParsedURL  *url.URL
	Rel        string
	Text       string
	External   bool
	NoFollow   bool
	Sponsored  bool
	UGC        bool
	StatusCode int
}
