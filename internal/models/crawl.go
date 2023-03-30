package models

import (
	"database/sql"
	"time"
)

type Crawl struct {
	Id                    int64
	ProjectId             int64
	URL                   string
	Start                 time.Time
	End                   sql.NullTime
	TotalIssues           int
	TotalURLs             int
	IssuesEnd             sql.NullTime
	CriticalIssues        int
	AlertIssues           int
	WarningIssues         int
	BlockedByRobotstxt    int // URLs blocked by robots.txt
	Noindex               int // URLS with noindex attribute
	SitemapExists         bool
	RobotstxtExists       bool
	InternalFollowLinks   int
	InternalNoFollowLinks int
	ExternalFollowLinks   int
	ExternalNoFollowLinks int
	SponsoredLinks        int
	UGCLinks              int
}
