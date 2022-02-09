package main

import (
	"database/sql"
	"time"
)

type Crawl struct {
	Id          int
	ProjectId   int
	URL         string
	Start       time.Time
	End         sql.NullTime
	TotalIssues int
	TotalURLs   int
}

func (c Crawl) TotalTime() time.Duration {
	if c.End.Valid {
		return c.End.Time.Sub(c.Start)
	}

	return 0
}
