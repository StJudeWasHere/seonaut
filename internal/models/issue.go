package models

type Issue struct {
	PageReportId int64
	CrawlId      int64
	ErrorType    int
}

type IssueGroup struct {
	ErrorType string
	Priority  int
	Count     int
}

type IssueCount struct {
	CriticalIssues []IssueGroup
	AlertIssues    []IssueGroup
	WarningIssues  []IssueGroup
	PassedIssues   []IssueGroup
}
