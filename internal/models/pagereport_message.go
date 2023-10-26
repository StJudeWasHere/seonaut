package models

import (
	"net/http"

	"golang.org/x/net/html"
)

type PageReportMessage struct {
	PageReport *PageReport
	HtmlNode   *html.Node
	Header     *http.Header
	Crawled    int
	Discovered int
}
