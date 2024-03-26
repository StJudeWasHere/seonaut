package models

import (
	"net/http"

	"golang.org/x/net/html"
)

// The PageIssueReporter struct contains a callback function and an error type.
// Each PageIssueReporter callback will be called and an issue will be created if it returns true.
type PageIssueReporter struct {
	Callback  func(*PageReport, *html.Node, *http.Header) bool
	ErrorType int
}

// The MultipageIssueReporter struct contains an int64 stream, which corresponds to the PageReport id,
// and an error type. Each MultipageIssueReporter will be called and an issue will be created for each
// PageReport which id is received through the channel.
type MultipageIssueReporter struct {
	Pstream   <-chan int64
	ErrorType int
}

type MultipageCallback func(c *Crawl) *MultipageIssueReporter
