package app

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/mnlg/lenkrr/internal/user"
)

func (app *App) serveCrawl(user *user.User, w http.ResponseWriter, r *http.Request) {
	qpid, ok := r.URL.Query()["pid"]
	if !ok || len(qpid) < 1 {
		log.Println("serveCrawl: pid parameter is missing")
		return
	}

	pid, err := strconv.Atoi(qpid[0])
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	p, err := app.datastore.findProjectById(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	go func() {
		log.Printf("Crawling %s\n", p.URL)

		start := time.Now()
		cid := startCrawler(p, app.config.CrawlerAgent, user.Advanced, app.datastore, app.sanitizer)

		log.Printf("Done crawling %s in %s\n", p.URL, time.Since(start))
		log.Printf("Creating issues for %s and crawl id %d\n", p.URL, cid)

		rm := ReportManager{}

		rm.addReporter(app.datastore.Find30xPageReports, Error30x)
		rm.addReporter(app.datastore.Find40xPageReports, Error40x)
		rm.addReporter(app.datastore.Find50xPageReports, Error50x)
		rm.addReporter(app.datastore.FindPageReportsWithDuplicatedTitle, ErrorDuplicatedTitle)
		rm.addReporter(app.datastore.FindPageReportsWithDuplicatedTitle, ErrorDuplicatedDescription)
		rm.addReporter(app.datastore.FindPageReportsWithEmptyTitle, ErrorEmptyTitle)
		rm.addReporter(app.datastore.FindPageReportsWithShortTitle, ErrorShortTitle)
		rm.addReporter(app.datastore.FindPageReportsWithLongTitle, ErrorLongTitle)
		rm.addReporter(app.datastore.FindPageReportsWithEmptyDescription, ErrorEmptyDescription)
		rm.addReporter(app.datastore.FindPageReportsWithShortDescription, ErrorShortDescription)
		rm.addReporter(app.datastore.FindPageReportsWithLongDescription, ErrorLongDescription)
		rm.addReporter(app.datastore.FindPageReportsWithLittleContent, ErrorLittleContent)
		rm.addReporter(app.datastore.FindImagesWithNoAlt, ErrorImagesWithNoAlt)
		rm.addReporter(app.datastore.findRedirectChains, ErrorRedirectChain)
		rm.addReporter(app.datastore.FindPageReportsWithoutH1, ErrorNoH1)
		rm.addReporter(app.datastore.FindPageReportsWithNoLangAttr, ErrorNoLang)
		rm.addReporter(app.datastore.FindPageReportsWithHTTPLinks, ErrorHTTPLinks)
		rm.addReporter(app.datastore.FindMissingHrelangReturnLinks, ErrorHreflangsReturnLink)
		rm.addReporter(app.datastore.tooManyLinks, ErrorTooManyLinks)
		rm.addReporter(app.datastore.internalNoFollowLinks, ErrorInternalNoFollow)
		rm.addReporter(app.datastore.findExternalLinkWitoutNoFollow, ErrorExternalWithoutNoFollow)
		rm.addReporter(app.datastore.findCanonicalizedToNonCanonical, ErrorCanonicalizedToNonCanonical)
		rm.addReporter(app.datastore.findCanonicalizedToNonCanonical, ErrorRedirectLoop)
		rm.addReporter(app.datastore.findNotValidHeadingsOrder, ErrorNotValidHeadings)

		issues := rm.createIssues(cid)
		app.datastore.saveIssues(issues, cid)

		totalIssues := len(issues)

		app.datastore.saveEndIssues(cid, time.Now(), totalIssues)

		log.Printf("Done creating issues for %s...\n", p.URL)
		log.Printf("Deleting previous crawl data for %s\n", p.URL)
		app.datastore.DeletePreviousCrawl(p.Id)
		log.Printf("Deleted previous crawl done for %s\n", p.URL)
	}()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
