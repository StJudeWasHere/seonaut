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
	"github.com/stjudewashere/seonaut/internal/report_manager/sql_reporters"
	"github.com/stjudewashere/seonaut/internal/user"
)

func main() {
	var fname string
	var path string

	flag.StringVar(&fname, "c", "config", "Specify config filename. Default is config")
	flag.StringVar(&path, "p", ".", "Specify config path. Default is current directory")
	flag.Parse()

	// Load config file.
	config, err := config.NewConfig(path, fname)
	if err != nil {
		log.Fatalf("Error loading config: %v\n", err)
	}

	// Create the sql database connection.
	db, err := datastore.SqlConnect(config.DB)
	if err != nil {
		log.Fatalf("Error creating new datastore: %v\n", err)
	}

	// Create database data store.
	ds := datastore.NewDataStore(db)

	// Run database migrations.
	err = ds.Migrate()
	if err != nil {
		log.Fatalf("Error running migrations: %v\n", err)
	}

	// Delete any unfinished crawls.
	unfinishedCrawls := ds.DeleteUnfinishedCrawls()
	log.Printf("Deleted %d unfinished crawls.", unfinishedCrawls)

	// Build services.
	broker := pubsub.New()
	cache := cache.NewMemCache()

	issueService := issue.NewService(ds, cache)
	reportService := report.NewService(ds, cache)

	cacheManager := cache_manager.New()
	cacheManager.AddCrawlCacheHandler(issueService)
	cacheManager.AddCrawlCacheHandler(reportService)

	reportManager := report_manager.NewReportManager(ds)
	for _, r := range reporters.GetAllReporters() {
		reportManager.AddPageReporter(r)
	}

	// Create the sql multipage reporters and add them all to the reporterManager.
	sqlReporters := sql_reporters.NewSqlReporter(db)
	for _, r := range sqlReporters.GetAllReporters() {
		reportManager.AddMultipageReporter(r)
	}

	// Start HTTP server.
	services := &http.Services{
		UserService:        user.NewService(ds),
		ProjectService:     project.NewService(ds, cacheManager),
		CrawlerService:     crawler.NewService(ds, broker, config.Crawler, cacheManager, reportManager, issueService),
		IssueService:       issueService,
		ReportService:      reportService,
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
