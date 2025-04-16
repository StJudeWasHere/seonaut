package models

type PageReportView struct {
	PageReport PageReport
	ErrorTypes []string
	InLinks    []InternalLink
	Redirects  []PageReport
	Paginator  Paginator
}
