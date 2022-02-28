package report

type ReportStore interface {
	FindPageReportById(int) PageReport
	FindErrorTypesByPage(int, int) []string
	FindInLinks(string, int) []PageReport
	FindPageReportsRedirectingToURL(string, int) []PageReport
}

type ReportService struct {
	store ReportStore
}

type PageReportView struct {
	PageReport PageReport
	ErrorTypes []string
	InLinks    []PageReport
	Redirects  []PageReport
}

func NewService(store ReportStore) *ReportService {
	return &ReportService{
		store: store,
	}
}

func (s *ReportService) GetPageReport(rid, crawlId int, tab string) *PageReportView {
	v := &PageReportView{
		PageReport: s.store.FindPageReportById(rid),
		ErrorTypes: s.store.FindErrorTypesByPage(rid, crawlId),
	}

	switch tab {
	case "inlinks":
		v.InLinks = s.store.FindInLinks(v.PageReport.URL, crawlId)
	case "redirections":
		v.Redirects = s.store.FindPageReportsRedirectingToURL(v.PageReport.URL, crawlId)
	}

	return v
}
