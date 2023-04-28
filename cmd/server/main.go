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
	"github.com/stjudewashere/seonaut/internal/report_manager/reporters"
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

	reportManager := newReportManager(ds, cacheManager)

	// Start HTTP server.
	services := &http.Services{
		UserService:        user.NewService(ds),
		ProjectService:     project.NewService(ds, cacheManager),
		CrawlerService:     crawler.NewService(ds, broker, config.Crawler, cacheManager, reportManager),
		IssueService:       issueService,
		ReportService:      reportService,
		ReportManager:      reportManager,
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

	rm.AddReporter(ds.FindPageReportsWithDuplicatedTitle, report_manager.ErrorDuplicatedTitle)
	rm.AddReporter(ds.FindPageReportsWithDuplicatedDescription, report_manager.ErrorDuplicatedDescription)
	rm.AddReporter(ds.FindRedirectChains, report_manager.ErrorRedirectChain)
	rm.AddReporter(ds.FindMissingHrelangReturnLinks, report_manager.ErrorHreflangsReturnLink)
	rm.AddReporter(ds.FindCanonicalizedToNonCanonical, report_manager.ErrorCanonicalizedToNonCanonical)
	rm.AddReporter(ds.FindRedirectLoops, report_manager.ErrorRedirectLoop)
	rm.AddReporter(ds.FindHreflangsToNonCanonical, report_manager.HreflangToNonCanonical)
	rm.AddReporter(ds.FindOrphanPages, report_manager.ErrorOrphan)
	rm.AddReporter(ds.FindIncomingIndexNoIndex, report_manager.IncomingFollowNofollow)
	rm.AddReporter(ds.HreflangNoindexable, report_manager.HreflangNoindexable)
	rm.AddReporter(ds.InternalNoFollowIndexableLinks, report_manager.ErrorInternalNoFollowIndexable)

	// Image issues
	rm.AddPageReporter(reporters.NoAltText, report_manager.ErrorImagesWithNoAlt)

	// Link issues
	rm.AddPageReporter(reporters.HTTPLinks, report_manager.ErrorHTTPLinks)
	rm.AddPageReporter(reporters.TooManyLinks, report_manager.ErrorTooManyLinks)
	rm.AddPageReporter(reporters.InternalNoFollowLinks, report_manager.ErrorInternalNoFollow)
	rm.AddPageReporter(reporters.ExternalLinkWitoutNoFollow, report_manager.ErrorExternalWithoutNoFollow)

	// Status code issues
	rm.AddPageReporter(reporters.Status30x, report_manager.Error30x)
	rm.AddPageReporter(reporters.Status40x, report_manager.Error40x)
	rm.AddPageReporter(reporters.Status50x, report_manager.Error40x)

	// Title issues
	rm.AddPageReporter(reporters.EmptyTitle, report_manager.ErrorEmptyTitle)
	rm.AddPageReporter(reporters.ShortTitle, report_manager.ErrorShortTitle)
	rm.AddPageReporter(reporters.LongTitle, report_manager.ErrorLongTitle)

	// Description issues
	rm.AddPageReporter(reporters.EmptyDescription, report_manager.ErrorEmptyDescription)
	rm.AddPageReporter(reporters.ShortDescription, report_manager.ErrorShortDescription)
	rm.AddPageReporter(reporters.LongDescription, report_manager.ErrorLongDescription)

	// Indexability issues
	rm.AddPageReporter(reporters.NoIndexable, report_manager.ErrorNoIndexable)
	rm.AddPageReporter(reporters.BlockedByRobotstxt, report_manager.ErrorBlocked)
	rm.AddPageReporter(reporters.NoIndexInSitemap, report_manager.SitemapNoIndex)
	rm.AddPageReporter(reporters.SitemapAndBlocked, report_manager.SitemapBlocked)
	rm.AddPageReporter(reporters.NonCanonicalInSitemap, report_manager.SitemapNonCanonical)

	// Language issues
	rm.AddPageReporter(reporters.InvalidLang, report_manager.InvalidLanguage)
	rm.AddPageReporter(reporters.MissingLang, report_manager.ErrorNoLang)

	// Content issues
	rm.AddPageReporter(reporters.LittleContent, report_manager.ErrorLittleContent)

	// Heading Issues
	rm.AddPageReporter(reporters.NoH1, report_manager.ErrorNoH1)
	rm.AddPageReporter(reporters.ValidHeadingsOrder, report_manager.ErrorNotValidHeadings)

	return rm
}
