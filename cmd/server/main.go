package main

import (
	"flag"
	"log"

	"github.com/stjudewashere/seonaut/internal/cache"
	"github.com/stjudewashere/seonaut/internal/cache_manager"
	"github.com/stjudewashere/seonaut/internal/config"
	"github.com/stjudewashere/seonaut/internal/crawler"
	"github.com/stjudewashere/seonaut/internal/datastore"
	"github.com/stjudewashere/seonaut/internal/export"
	"github.com/stjudewashere/seonaut/internal/http"
	"github.com/stjudewashere/seonaut/internal/issue"
	"github.com/stjudewashere/seonaut/internal/project"
	"github.com/stjudewashere/seonaut/internal/projectview"
	"github.com/stjudewashere/seonaut/internal/pubsub"
	"github.com/stjudewashere/seonaut/internal/report"
	"github.com/stjudewashere/seonaut/internal/report_manager"
	"github.com/stjudewashere/seonaut/internal/user"
)

func main() {
	var fname string
	var path string

	flag.StringVar(&fname, "c", "config", "Specify config filename. Default is config")
	flag.StringVar(&path, "p", ".", "Specify config path. Default is current directory")
	flag.Parse()

	config, err := config.NewConfig(path, fname)
	if err != nil {
		log.Fatalf("Error loading config: %v\n", err)
	}

	ds, err := datastore.NewDataStore(config.DB)
	if err != nil {
		log.Fatalf("Error creating new datastore: %v\n", err)
	}

	err = ds.Migrate()
	if err != nil {
		log.Fatalf("Error running migrations: %v\n", err)
	}

	broker := pubsub.New()
	cache := cache.New(config.Cache)

	issueService := issue.NewService(ds, cache)
	reportService := report.NewService(ds, cache)

	cacheManager := cache_manager.New()
	cacheManager.AddCrawlCacheHandler(issueService)
	cacheManager.AddCrawlCacheHandler(reportService)

	services := &http.Services{
		UserService:        user.NewService(ds),
		ProjectService:     project.NewService(ds, cacheManager),
		CrawlerService:     crawler.NewService(ds, broker, config.Crawler, cacheManager),
		IssueService:       issueService,
		ReportService:      reportService,
		ReportManager:      newReportManager(ds, cacheManager),
		ProjectViewService: projectview.NewService(ds),
		PubSubBroker:       broker,
		ExportService:      export.NewExporter(ds),
	}

	server := http.NewApp(
		config.HTTPServer,
		services,
	)

	server.Run()
}

// Create a new ReportManager with all available issue reporters.
func newReportManager(ds *datastore.Datastore, cm *cache_manager.CacheManager) *report_manager.ReportManager {
	rm := report_manager.NewReportManager(ds, cm)

	rm.AddReporter(ds.Find30xPageReports, report_manager.Error30x)
	rm.AddReporter(ds.Find40xPageReports, report_manager.Error40x)
	rm.AddReporter(ds.Find50xPageReports, report_manager.Error50x)
	rm.AddReporter(ds.FindPageReportsWithDuplicatedTitle, report_manager.ErrorDuplicatedTitle)
	rm.AddReporter(ds.FindPageReportsWithDuplicatedDescription, report_manager.ErrorDuplicatedDescription)
	rm.AddReporter(ds.FindPageReportsWithEmptyTitle, report_manager.ErrorEmptyTitle)
	rm.AddReporter(ds.FindPageReportsWithShortTitle, report_manager.ErrorShortTitle)
	rm.AddReporter(ds.FindPageReportsWithLongTitle, report_manager.ErrorLongTitle)
	rm.AddReporter(ds.FindPageReportsWithEmptyDescription, report_manager.ErrorEmptyDescription)
	rm.AddReporter(ds.FindPageReportsWithShortDescription, report_manager.ErrorShortDescription)
	rm.AddReporter(ds.FindPageReportsWithLongDescription, report_manager.ErrorLongDescription)
	rm.AddReporter(ds.FindPageReportsWithLittleContent, report_manager.ErrorLittleContent)
	rm.AddReporter(ds.FindImagesWithNoAlt, report_manager.ErrorImagesWithNoAlt)
	rm.AddReporter(ds.FindRedirectChains, report_manager.ErrorRedirectChain)
	rm.AddReporter(ds.FindPageReportsWithoutH1, report_manager.ErrorNoH1)
	rm.AddReporter(ds.FindPageReportsWithNoLangAttr, report_manager.ErrorNoLang)
	rm.AddReporter(ds.FindPageReportsWithHTTPLinks, report_manager.ErrorHTTPLinks)
	rm.AddReporter(ds.FindMissingHrelangReturnLinks, report_manager.ErrorHreflangsReturnLink)
	rm.AddReporter(ds.TooManyLinks, report_manager.ErrorTooManyLinks)
	rm.AddReporter(ds.InternalNoFollowLinks, report_manager.ErrorInternalNoFollow)
	rm.AddReporter(ds.FindExternalLinkWitoutNoFollow, report_manager.ErrorExternalWithoutNoFollow)
	rm.AddReporter(ds.FindCanonicalizedToNonCanonical, report_manager.ErrorCanonicalizedToNonCanonical)
	rm.AddReporter(ds.FindRedirectLoops, report_manager.ErrorRedirectLoop)
	rm.AddReporter(ds.FindNotValidHeadingsOrder, report_manager.ErrorNotValidHeadings)
	rm.AddReporter(ds.FindHreflangsToNonCanonical, report_manager.HreflangToNonCanonical)
	rm.AddReporter(ds.InternalNoFollowIndexableLinks, report_manager.ErrorInternalNoFollowIndexable)
	rm.AddReporter(ds.NoIndexable, report_manager.ErrorNoIndexable)
	rm.AddReporter(ds.HreflangNoindexable, report_manager.HreflangNoindexable)
	rm.AddReporter(ds.FindBlockedByRobotstxt, report_manager.ErrorBlocked)
	rm.AddReporter(ds.FindOrphanPages, report_manager.ErrorOrphan)
	rm.AddReporter(ds.FindNoIndexInSitemap, report_manager.SitemapNoIndex)
	rm.AddReporter(ds.FindBlockedInSitemap, report_manager.SitemapBlocked)
	rm.AddReporter(ds.FindNonCanonicalInSitemap, report_manager.SitemapNonCanonical)
	rm.AddReporter(ds.FindIncomingIndexNoIndex, report_manager.IncomingFollowNofollow)
	rm.AddReporter(ds.FindInvalidLang, report_manager.InvalidLanguage)

	return rm
}
