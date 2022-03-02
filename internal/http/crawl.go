package http

import (
	"log"
	"net/http"
	"strconv"

	"github.com/stjudewashere/seonaut/internal/user"
)

func (app *App) serveCrawl(user *user.User, w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		log.Printf("serveCrawl pid: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	p, err := app.projectService.FindProject(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	go func() {
		cid := app.crawlerService.StartCrawler(p, app.config.CrawlerAgent, app.sanitizer)

		log.Printf("Creating issues for %s and crawl id %d\n", p.URL, cid)
		app.reportManager.CreateIssues(cid)
		log.Printf("Done creating issues for %s...\n", p.URL)
	}()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
