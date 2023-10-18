package models

import (
	"net/http"

	"golang.org/x/net/html"
)

type PageReportMessage struct {
	PageReport *PageReport
	HtmlNode   *html.Node
	Response   *http.Response
	Crawled    int
	Discovered int
}
