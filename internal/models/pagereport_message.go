package models

type PageReportMessage struct {
	PageReport *PageReport
	Crawled    int
	Discovered int
}
