package models

import (
	"time"
)

type Project struct {
	Id                 int64
	URL                string
	Host               string
	IgnoreRobotsTxt    bool
	FollowNofollow     bool
	IncludeNoindex     bool
	Created            time.Time
	CrawlSitemap       bool
	AllowSubdomains    bool
	Deleting           bool
	BasicAuth          bool
	CheckExternalLinks bool
	Archive            bool
	UserAgent          string
	WebhookURL         string     // POST JSON crawl summary here after each crawl. Empty disables notifications.
	ScheduleInterval   string     // Recrawl interval: "", "hourly", "daily" or "weekly".
	NextRun            *time.Time // Next time the scheduler will start a crawl for this project.
	Scheduled          bool       // Transient: set by the scheduler when it starts a run. Not persisted.
}
