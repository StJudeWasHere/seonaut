package models

import (
	"time"
)

type Crawl struct {
	Id        int64
	ProjectId int64
	Crawling  bool

	URL                   string
	Start                 time.Time
	End                   time.Time
	TotalIssues           int
	TotalURLs             int
	IssuesEnd             time.Time
	CriticalIssues        int
	AlertIssues           int
	WarningIssues         int
	BlockedByRobotstxt    int // URLs blocked by robots.txt
	Noindex               int // URLS with noindex attribute
	SitemapExists         bool
	SitemapIsBlocked      bool
	RobotstxtExists       bool
	InternalFollowLinks   int
	InternalNoFollowLinks int
	ExternalFollowLinks   int
	ExternalNoFollowLinks int
	SponsoredLinks        int
	UGCLinks              int
}
