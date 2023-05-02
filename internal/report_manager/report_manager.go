package report_manager

import (
	"sync"
	"time"

	"github.com/stjudewashere/seonaut/internal/cache_manager"
	"github.com/stjudewashere/seonaut/internal/issue"
	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporters"
)

type ReportManager struct {
	store              ReportManagerStore
	pageCallbacks      []*reporters.PageIssueReporter
	multipageCallbacks []reporters.MultipageCallback
	cacheManager       *cache_manager.CacheManager
}

type ReportManagerStore interface {
	SaveIssues(<-chan *issue.Issue)
	SaveEndIssues(int64, time.Time)
	PageReportsQuery(query string, args ...interface{}) <-chan *models.PageReport
}

// Create a new ReportManager with no issue reporters.
func NewReportManager(s ReportManagerStore, cm *cache_manager.CacheManager) *ReportManager {
	return &ReportManager{
		store:        s,
		cacheManager: cm,
	}
}

// Create a new ReportManager with all available issue reporters.
// It includes both, page issue reporters and multi-page issue reporters.
func NewFullReportManager(s ReportManagerStore, cm *cache_manager.CacheManager) *ReportManager {
	rm := NewReportManager(s, cm)

	// Add status code issue reporters
	rm.AddPageReporter(reporters.NewStatus30xReporter())
	rm.AddPageReporter(reporters.NewStatus40xReporter())
	rm.AddPageReporter(reporters.NewStatus50xReporter())

	rm.AddMultipageReporter(reporters.RedirectChainsReporter)
	rm.AddMultipageReporter(reporters.RedirectLoopsReporter)

	// Add title issue reporters
	rm.AddPageReporter(reporters.NewEmptyTitleReporter())
	rm.AddPageReporter(reporters.NewShortTitleReporter())
	rm.AddPageReporter(reporters.NewShortTitleReporter())

	rm.AddMultipageReporter(reporters.DuplicatedTitleReporter)

	// Add description issue reporters
	rm.AddPageReporter(reporters.NewEmptyDescriptionReporter())
	rm.AddPageReporter(reporters.NewShortDescriptionReporter())
	rm.AddPageReporter(reporters.NewLongDescriptionReporter())

	rm.AddMultipageReporter(reporters.DuplicatedDescriptionReporter)

	// Add indexability issue reporters
	rm.AddPageReporter(reporters.NewNoIndexableReporter())
	rm.AddPageReporter(reporters.NewBlockedByRobotstxtReporter())
	rm.AddPageReporter(reporters.NewNoIndexInSitemapReporter())
	rm.AddPageReporter(reporters.NewSitemapAndBlockedReporter())
	rm.AddPageReporter(reporters.NewNonCanonicalInSitemapReporter())

	// Add link issue reporters
	rm.AddPageReporter(reporters.NewTooManyLinksReporter())
	rm.AddPageReporter(reporters.NewInternalNoFollowLinksReporter())
	rm.AddPageReporter(reporters.NewExternalLinkWitoutNoFollowReporter())
	rm.AddPageReporter(reporters.NewHTTPLinksReporter())

	rm.AddMultipageReporter(reporters.OrphanPagesReporter)
	rm.AddMultipageReporter(reporters.NoFollowIndexableReporter)
	rm.AddMultipageReporter(reporters.FollowNoFollowReporter)

	// Add image issue reporters
	rm.AddPageReporter(reporters.NewAltTextReporter())

	// Add language issue reporters
	rm.AddPageReporter(reporters.NewInvalidLangReporter())
	rm.AddPageReporter(reporters.NewMissingLangReporter())

	// Add hreflang reporters
	rm.AddMultipageReporter(reporters.MissingHrelangReturnLinks)
	rm.AddMultipageReporter(reporters.HreflangsToNonCanonical)
	rm.AddMultipageReporter(reporters.HreflangNoindexable)

	// Add heading issue reporters
	rm.AddPageReporter(reporters.NewNoH1Reporter())
	rm.AddPageReporter(reporters.NewValidHeadingsOrderReporter())

	// Add content issue reporters
	rm.AddPageReporter(reporters.NewLittleContentReporter())

	// Add canonical issue reporters
	rm.AddMultipageReporter(reporters.CanonicalizedToNonCanonical)

	return rm
}

// Add a multi-page issue reporter to the ReportManager. Multi-page reporters are used to detect
// issues that affect multiple pages. It will be used when creating the multi page issues once all
// the pages have been crawled.
func (r *ReportManager) AddMultipageReporter(reporter reporters.MultipageCallback) {
	r.multipageCallbacks = append(r.multipageCallbacks, reporter)
}

// Add an page issue reporter to the ReportManager.
// It will be used to create issues on each crawled page.
func (r *ReportManager) AddPageReporter(reporter *reporters.PageIssueReporter) {
	r.pageCallbacks = append(r.pageCallbacks, reporter)
}

// CreateIssues uses the Reporters to create and save issues found in a crawl.
func (r *ReportManager) CreateIssues(crawl *models.Crawl) {
	issueCount := 0
	iStream := make(chan *issue.Issue)

	wg := new(sync.WaitGroup)
	wg.Add(1)

	go func() {
		r.store.SaveIssues(iStream)
		wg.Done()
	}()

	for _, callback := range r.multipageCallbacks {
		reporter := callback(crawl)
		for p := range r.store.PageReportsQuery(reporter.Query, reporter.Parameters...) {
			iStream <- &issue.Issue{
				PageReportId: p.Id,
				CrawlId:      crawl.Id,
				ErrorType:    reporter.ErrorType,
			}

			issueCount++
		}
	}

	close(iStream)

	wg.Wait()

	r.store.SaveEndIssues(crawl.Id, time.Now())
	r.cacheManager.BuildCrawlCache(crawl)
}

// CreatePageIssues loops the page reporters calling the callback function
// and creating the issues found in the PageReport.
func (r *ReportManager) CreatePageIssues(p *models.PageReport, crawl *models.Crawl) {
	iStream := make(chan *issue.Issue)
	wg := new(sync.WaitGroup)
	wg.Add(1)

	go func() {
		r.store.SaveIssues(iStream)
		wg.Done()
	}()

	for _, c := range r.pageCallbacks {
		if c.Callback(p) == true {
			iStream <- &issue.Issue{
				PageReportId: p.Id,
				CrawlId:      crawl.Id,
				ErrorType:    c.ErrorType,
			}
		}
	}

	close(iStream)

	wg.Wait()
}
