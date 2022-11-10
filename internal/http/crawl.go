package http

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/stjudewashere/seonaut/internal/crawler"
	"github.com/stjudewashere/seonaut/internal/project"
	"github.com/stjudewashere/seonaut/internal/pubsub"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

// Start crawling a project
func (app *App) serveCrawl(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		log.Printf("serveCrawl pid: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	user, ok := app.userService.GetUserFromContext(r.Context())
	if ok == false {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
	}

	p, err := app.projectService.FindProject(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if p.BasicAuth == true {
		http.Redirect(w, r, "/crawl-auth?id="+strconv.Itoa(pid), http.StatusSeeOther)
		return
	}

	go app.startCrawler(p)

	http.Redirect(w, r, "/crawl-live?pid="+strconv.Itoa(pid), http.StatusSeeOther)
}

func (app *App) serveCrawlAuth(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		log.Printf("serveCrawl pid: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	user, ok := app.userService.GetUserFromContext(r.Context())
	if ok == false {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
	}

	p, err := app.projectService.FindProject(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Printf("serveCrawlAuth ParseForm: %v\n", err)
			http.Redirect(w, r, "/crawl-auth", http.StatusSeeOther)

			return
		}

		p.BasicAuth = true
		p.AuthUser = r.FormValue("username")
		p.AuthPass = r.FormValue("password")

		go app.startCrawler(p)

		http.Redirect(w, r, "/crawl-live?pid="+strconv.Itoa(pid), http.StatusSeeOther)
	}

	pageView := &PageView{
		PageTitle: "CRAWL_AUTH_VIEW",
		Data:      struct{ Project project.Project }{Project: p},
	}

	app.renderer.RenderTemplate(w, "crawl_auth", pageView)
}

func (app *App) startCrawler(p project.Project) {
	log.Printf("Crawling %s\n", p.URL)
	crawl, err := app.crawlerService.StartCrawler(p)
	if err != nil {
		log.Printf("StartCrawler: %s %v\n", p.URL, err)
		return
	}

	log.Printf("Crawled %d pages at %s\n", crawl.TotalURLs, p.URL)

	app.pubsubBroker.Publish(fmt.Sprintf("crawl-%d", p.Id), &pubsub.Message{Name: "IssuesInit"})
	app.reportManager.CreateIssues(crawl.Id)
	app.issueService.SaveCrawlIssuesCount(crawl.Id)
	app.pubsubBroker.Publish(fmt.Sprintf("crawl-%d", p.Id), &pubsub.Message{Name: "CrawlEnd"})
}

// Start crawling a project
func (app *App) serveCrawlLive(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	user, ok := app.userService.GetUserFromContext(r.Context())
	if ok == false {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
	}

	pv, err := app.projectViewService.GetProjectView(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if pv.Crawl.IssuesEnd.Valid {
		http.Redirect(w, r, "/dashboard?pid="+strconv.Itoa(pid), http.StatusSeeOther)
		return
	}

	configURL, err := url.Parse(app.config.URL)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	v := &PageView{
		Data: struct {
			Project project.Project
			Secure  bool
		}{
			Project: pv.Project,
			Secure:  configURL.Scheme == "https",
		},
		User:      *user,
		PageTitle: "CRAWL_LIVE",
	}

	app.renderer.RenderTemplate(w, "crawl_live", v)
}

// Handle the crawler's websocket connection
func (app *App) serveCrawlWs(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, ok := app.userService.GetUserFromContext(r.Context())
	if ok == false {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	p, err := app.projectService.FindProject(pid, user.Id)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Upgrade connection
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			return origin == app.config.URL
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error { conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	connLock := &sync.RWMutex{}

	subscriber := app.pubsubBroker.NewSubscriber(fmt.Sprintf("crawl-%d", p.Id), func(i *pubsub.Message) error {
		pubsubMessage := i
		wsMessage := struct {
			Name string
			Data interface{}
		}{
			Name: pubsubMessage.Name,
		}

		if pubsubMessage.Name == "PageReport" {
			msg := pubsubMessage.Data.(*crawler.PageReportMessage)
			wsMessage.Data = struct {
				StatusCode int
				URL        string
				Crawled    int
				Discovered int
			}{
				StatusCode: msg.PageReport.StatusCode,
				URL:        msg.PageReport.URL,
				Crawled:    msg.Crawled,
				Discovered: msg.Discovered,
			}
		}

		connLock.Lock()
		conn.SetWriteDeadline(time.Now().Add(writeWait))
		err := conn.WriteJSON(wsMessage)
		connLock.Unlock()

		return err
	})
	defer app.pubsubBroker.Unsubscribe(subscriber)

	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				connLock.Lock()
				conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
				connLock.Unlock()
			}
		}
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			conn.Close()
			break
		}
	}
}
