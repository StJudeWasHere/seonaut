package report_manager

import (
	"sync"
	"time"

	"github.com/stjudewashere/seonaut/internal/cache_manager"
	"github.com/stjudewashere/seonaut/internal/issue"
	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporters"
)

type Reporter func(reporters.DatabaseReporter, *models.Crawl) <-chan *models.PageReport

type IssueCallback struct {
	Callback  Reporter
	ErrorType int
}

type ReportManager struct {
	store         ReportManagerStore
	callbacks     []IssueCallback
	pageCallbacks []*reporters.PageIssueReporter
	cacheManager  *cache_manager.CacheManager
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
func NewFullReportManager(s ReportManagerStore, cm *cache_manager.CacheManager) *ReportManager {
	rm := NewReportManager(s, cm)

	// Whole site issues
	rm.AddReporter(reporters.FindPageReportsWithDuplicatedTitle, reporters.ErrorDuplicatedTitle)
	rm.AddReporter(reporters.FindPageReportsWithDuplicatedDescription, reporters.ErrorDuplicatedDescription)
	rm.AddReporter(reporters.FindRedirectChains, reporters.ErrorRedirectChain)
	rm.AddReporter(reporters.FindMissingHrelangReturnLinks, reporters.ErrorHreflangsReturnLink)
	rm.AddReporter(reporters.FindCanonicalizedToNonCanonical, reporters.ErrorCanonicalizedToNonCanonical)
	rm.AddReporter(reporters.FindRedirectLoops, reporters.ErrorRedirectLoop)
	rm.AddReporter(reporters.FindHreflangsToNonCanonical, reporters.HreflangToNonCanonical)
	rm.AddReporter(reporters.FindOrphanPages, reporters.ErrorOrphan)
	rm.AddReporter(reporters.FindIncomingIndexNoIndex, reporters.IncomingFollowNofollow)
	rm.AddReporter(reporters.FindHreflangNoindexable, reporters.HreflangNoindexable)
	rm.AddReporter(reporters.InternalNoFollowIndexableLinks, reporters.ErrorInternalNoFollowIndexable)

	// Image issues
	rm.AddPageReporter(reporters.NewAltTextReporter())

	// Link issues
	rm.AddPageReporter(reporters.NewTooManyLinksReporter())
	rm.AddPageReporter(reporters.NewInternalNoFollowLinksReporter())
	rm.AddPageReporter(reporters.NewExternalLinkWitoutNoFollowReporter())
	rm.AddPageReporter(reporters.NewHTTPLinksReporter())

	// Status code issues
	rm.AddPageReporter(reporters.NewStatus30xReporter())
	rm.AddPageReporter(reporters.NewStatus40xReporter())
	rm.AddPageReporter(reporters.NewStatus50xReporter())

	// Title issues
	rm.AddPageReporter(reporters.NewEmptyTitleReporter())
	rm.AddPageReporter(reporters.NewShortTitleReporter())
	rm.AddPageReporter(reporters.NewShortTitleReporter())

	// Description issues
	rm.AddPageReporter(reporters.NewEmptyDescriptionReporter())
	rm.AddPageReporter(reporters.NewShortDescriptionReporter())
	rm.AddPageReporter(reporters.NewLongDescriptionReporter())

	// Indexability issues
	rm.AddPageReporter(reporters.NewNoIndexableReporter())
	rm.AddPageReporter(reporters.NewBlockedByRobotstxtReporter())
	rm.AddPageReporter(reporters.NewNoIndexInSitemapReporter())
	rm.AddPageReporter(reporters.NewSitemapAndBlockedReporter())
	rm.AddPageReporter(reporters.NewNonCanonicalInSitemapReporter())

	// Language issues
	rm.AddPageReporter(reporters.NewInvalidLangReporter())
	rm.AddPageReporter(reporters.NewMissingLangReporter())

	// Content issues
	rm.AddPageReporter(reporters.NewLittleContentReporter())

	// Heading issues
	rm.AddPageReporter(reporters.NewNoH1Reporter())
	rm.AddPageReporter(reporters.NewValidHeadingsOrderReporter())

	return rm
}

// Add an issue reporter to the ReportManager.
// It will be used when creating the issues.
func (r *ReportManager) AddReporter(c Reporter, t int) {
	r.callbacks = append(r.callbacks, IssueCallback{Callback: c, ErrorType: t})
}

// Add an issue page reporter to the ReportManager.
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

	for _, c := range r.callbacks {
		for p := range c.Callback(r.store, crawl) {
			i := &issue.Issue{
				PageReportId: p.Id,
				CrawlId:      crawl.Id,
				ErrorType:    c.ErrorType,
			}

			iStream <- i

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
