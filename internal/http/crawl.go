package http

import (
	"log"
	"net/http"
	"strconv"

	"github.com/stjudewashere/seonaut/internal/user"
)

func (app *App) serveCrawl(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		log.Printf("serveCrawl pid: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	c := r.Context().Value("user")
	user, ok := c.(*user.User)
	if ok == false {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
	}

	p, err := app.projectService.FindProject(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	go func() {
		log.Printf("Crawling %s\n", p.URL)
		crawl, err := app.crawlerService.StartCrawler(p)
		if err != nil {
			log.Printf("StartCrawler: %s %v\n", p.URL, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("Crawled %d pages at %s\n", crawl.TotalURLs, p.URL)

		app.reportManager.CreateIssues(crawl.Id)
	}()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
