package report_manager

import (
	"sync"
	"time"

	"github.com/stjudewashere/seonaut/internal/cache_manager"
	"github.com/stjudewashere/seonaut/internal/issue"
	"github.com/stjudewashere/seonaut/internal/models"
)

const (
	Error30x                         = iota + 1 // HTTP redirect
	Error40x                                    // HTTP not found
	Error50x                                    // HTTP internal error
	ErrorDuplicatedTitle                        // Duplicate title
	ErrorDuplicatedDescription                  // Duplicate description
	ErrorEmptyTitle                             // Missing or empty title
	ErrorShortTitle                             // Page title is too short
	ErrorLongTitle                              // Page title is too long
	ErrorEmptyDescription                       // Missing or empty meta description
	ErrorShortDescription                       // Meta description is too short
	ErrorLongDescription                        // Meta description is too long
	ErrorLittleContent                          // Not enough content
	ErrorImagesWithNoAlt                        // Images with no alt attribute
	ErrorRedirectChain                          // Redirect chain
	ErrorNoH1                                   // Missing or empy H1 tag
	ErrorNoLang                                 // Missing or empty html lang attribute
	ErrorHTTPLinks                              // Links using insecure http schema
	ErrorHreflangsReturnLink                    // Hreflang is not bidirectional
	ErrorTooManyLinks                           // Page contains too many links
	ErrorInternalNoFollow                       // Page has internal links with nofollow attribute
	ErrorExternalWithoutNoFollow                // Page has external follow links
	ErrorCanonicalizedToNonCanonical            // Page canonicalized to a non canonical page
	ErrorRedirectLoop                           // Redirect loop
	ErrorNotValidHeadings                       // H1-H6 tags have wrong order
	HreflangToNonCanonical                      // Hreflang to non canonical page
	ErrorInternalNoFollowIndexable              // Nofollow links to indexable pages
	ErrorNoIndexable                            // Page using the noindex attribute
	HreflangNoindexable                         // Hreflang to a non indexable page
	ErrorBlocked                                // Blocked by robots.txt
	ErrorOrphan                                 // Orphan pages
	SitemapNoIndex                              // No index pages included in the sitemap
	SitemapBlocked                              // Pages included in the sitemap that are blocked in robots.txt
	SitemapNonCanonical                         // Non canonical pages included in the sitemap
	IncomingFollowNofollow                      // Pages with index and noindex incoming links
	InvalidLanguage                             // Pages with invalid lang attribute
)

type Reporter func(int64) <-chan *models.PageReport
type PageReporter func(*models.PageReport) bool

type IssueCallback struct {
	Callback  Reporter
	ErrorType int
}
type PageIssueCallback struct {
	Callback  PageReporter
	ErrorType int
}

type ReportManager struct {
	store         ReportManagerStore
	callbacks     []IssueCallback
	pageCallbacks []PageIssueCallback
	cacheManager  *cache_manager.CacheManager
}

type ReportManagerStore interface {
	SaveIssues(<-chan *issue.Issue)
	SaveEndIssues(int64, time.Time)
}

func NewReportManager(s ReportManagerStore, cm *cache_manager.CacheManager) *ReportManager {
	return &ReportManager{
		store:        s,
		cacheManager: cm,
	}
}

// Add an issue reporter to the ReportManager.
// It will be used when creating the issues.
func (r *ReportManager) AddReporter(c Reporter, t int) {
	r.callbacks = append(r.callbacks, IssueCallback{Callback: c, ErrorType: t})
}

// Add an issue page reporter to the ReportManager.
// It will be used to create issues on each crawled page.
func (r *ReportManager) AddPageReporter(c PageReporter, t int) {
	r.pageCallbacks = append(r.pageCallbacks, PageIssueCallback{Callback: c, ErrorType: t})
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
		for p := range c.Callback(crawl.Id) {
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
		result := c.Callback(p)
		if result == true {
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
