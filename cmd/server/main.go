package main

import (
	"flag"
	"log"

	"github.com/stjudewashere/seonaut/internal/config"
	"github.com/stjudewashere/seonaut/internal/crawler"
	"github.com/stjudewashere/seonaut/internal/datastore"
	"github.com/stjudewashere/seonaut/internal/http"
	"github.com/stjudewashere/seonaut/internal/issue"
	"github.com/stjudewashere/seonaut/internal/project"
	"github.com/stjudewashere/seonaut/internal/report"
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

	services := &http.Services{
		UserService:    user.NewService(ds),
		ProjectService: project.NewService(ds),
		CrawlerService: crawler.NewService(ds, config.Crawler),
		IssueService:   issue.NewService(ds),
		ReportService:  report.NewService(ds),
		ReportManager:  newReportManager(ds),
	}

	server := http.NewApp(
		config.HTTPServer,
		services,
	)

	server.Run()
}

func newReportManager(ds *datastore.Datastore) *issue.ReportManager {
	rm := issue.NewReportManager(ds)

	rm.AddReporter(ds.Find30xPageReports, http.Error30x)
	rm.AddReporter(ds.Find40xPageReports, http.Error40x)
	rm.AddReporter(ds.Find50xPageReports, http.Error50x)
	rm.AddReporter(ds.FindPageReportsWithDuplicatedTitle, http.ErrorDuplicatedTitle)
	rm.AddReporter(ds.FindPageReportsWithDuplicatedTitle, http.ErrorDuplicatedDescription)
	rm.AddReporter(ds.FindPageReportsWithEmptyTitle, http.ErrorEmptyTitle)
	rm.AddReporter(ds.FindPageReportsWithShortTitle, http.ErrorShortTitle)
	rm.AddReporter(ds.FindPageReportsWithLongTitle, http.ErrorLongTitle)
	rm.AddReporter(ds.FindPageReportsWithEmptyDescription, http.ErrorEmptyDescription)
	rm.AddReporter(ds.FindPageReportsWithShortDescription, http.ErrorShortDescription)
	rm.AddReporter(ds.FindPageReportsWithLongDescription, http.ErrorLongDescription)
	rm.AddReporter(ds.FindPageReportsWithLittleContent, http.ErrorLittleContent)
	rm.AddReporter(ds.FindImagesWithNoAlt, http.ErrorImagesWithNoAlt)
	rm.AddReporter(ds.FindRedirectChains, http.ErrorRedirectChain)
	rm.AddReporter(ds.FindPageReportsWithoutH1, http.ErrorNoH1)
	rm.AddReporter(ds.FindPageReportsWithNoLangAttr, http.ErrorNoLang)
	rm.AddReporter(ds.FindPageReportsWithHTTPLinks, http.ErrorHTTPLinks)
	rm.AddReporter(ds.FindMissingHrelangReturnLinks, http.ErrorHreflangsReturnLink)
	rm.AddReporter(ds.TooManyLinks, http.ErrorTooManyLinks)
	rm.AddReporter(ds.InternalNoFollowLinks, http.ErrorInternalNoFollow)
	rm.AddReporter(ds.FindExternalLinkWitoutNoFollow, http.ErrorExternalWithoutNoFollow)
	rm.AddReporter(ds.FindCanonicalizedToNonCanonical, http.ErrorCanonicalizedToNonCanonical)
	rm.AddReporter(ds.FindCanonicalizedToNonCanonical, http.ErrorRedirectLoop)
	rm.AddReporter(ds.FindNotValidHeadingsOrder, http.ErrorNotValidHeadings)

	return rm
}
