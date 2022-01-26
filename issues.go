package main

const (
	Error30x                   = "ERROR_30x"
	Error40x                   = "ERROR_40x"
	Error50x                   = "ERROR_50x"
	ErrorDuplicatedTitle       = "ERROR_DUPLICATED_TITLE"
	ErrorDuplicatedDescription = "ERROR_DUPLICATED_DESCRIPTION"
	ErrorEmptyTitle            = "ERROR_EMPTY_TITLE"
	ErrorShortTitle            = "ERROR_SHORT_TITLE"
	ErrorLongTitle             = "ERROR_LONG_TITLE"
	ErrorEmptyDescription      = "ERROR_EMPTY_DESCRIPTION"
	ErrorShortDescription      = "ERROR_SHORT_DESCRIPTION"
	ErrorLongDescription       = "ERROR_LONG_DESCRIPTION"
	ErrorLittleContent         = "ERROR_LITTLE_CONTENT"
	ErrorImagesWithNoAlt       = "ERROR_IMAGES_NO_ALT"
	ErrorRedirectChain         = "ERROR_REDIRECT_CHAIN"
	ErrorNoH1                  = "ERROR_NO_H1"
	ErrorNoLang                = "ERROR_NO_LANG"
	ErrorHTTPLinks             = "ERROR_HTTP_LINKS"
	ErrorHreflangsReturnLink   = "ERROR_HREFLANG_RETURN"
)

type Issue struct {
	PageReportId int
	ErrorType    string
}

type IssueGroup struct {
	ErrorType string
	Count     int
}

type IssueCallback struct {
	Callback  func(int) []PageReport
	ErrorType string
}

type ReportManager struct {
	callbacks []IssueCallback
}

func NewReportManager() *ReportManager {
	r := ReportManager{}

	r.addReporter(Find30xPageReports, Error30x)
	r.addReporter(Find40xPageReports, Error40x)
	r.addReporter(Find50xPageReports, Error50x)
	r.addReporter(FindPageReportsWithDuplicatedTitle, ErrorDuplicatedTitle)
	r.addReporter(FindPageReportsWithDuplicatedTitle, ErrorDuplicatedDescription)
	r.addReporter(FindPageReportsWithEmptyTitle, ErrorEmptyTitle)
	r.addReporter(FindPageReportsWithShortTitle, ErrorShortTitle)
	r.addReporter(FindPageReportsWithLongTitle, ErrorLongTitle)
	r.addReporter(FindPageReportsWithEmptyDescription, ErrorEmptyDescription)
	r.addReporter(FindPageReportsWithShortDescription, ErrorShortDescription)
	r.addReporter(FindPageReportsWithLongDescription, ErrorLongDescription)
	r.addReporter(FindPageReportsWithLittleContent, ErrorLittleContent)
	r.addReporter(FindImagesWithNoAlt, ErrorImagesWithNoAlt)
	r.addReporter(findRedirectChains, ErrorRedirectChain)
	r.addReporter(FindPageReportsWithoutH1, ErrorNoH1)
	r.addReporter(FindPageReportsWithNoLangAttr, ErrorNoLang)
	r.addReporter(FindPageReportsWithHTTPLinks, ErrorHTTPLinks)
	r.addReporter(FindMissingHrelangReturnLinks, ErrorHreflangsReturnLink)

	return &r
}

func (r *ReportManager) addReporter(c func(int) []PageReport, t string) {
	r.callbacks = append(r.callbacks, IssueCallback{Callback: c, ErrorType: t})
}

func (r *ReportManager) createIssues(cid int) {
	var issues []Issue

	for _, c := range r.callbacks {
		for _, p := range c.Callback(cid) {
			i := Issue{
				PageReportId: p.Id,
				ErrorType:    c.ErrorType,
			}

			issues = append(issues, i)
		}
	}

	saveIssues(issues, cid)
}
