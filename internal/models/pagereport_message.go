package models

type PageReportMessage struct {
	StatusCode int
	Crawled    int
	URL        string
	Crawling   bool
	Discovered int
}
