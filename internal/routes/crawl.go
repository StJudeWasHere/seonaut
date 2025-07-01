package routes

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/services"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

type crawlHandler struct {
	*services.Container
}

// startHandler handles the crawling of a project.
// It expects a query parameter "pid" containing the project id to be crawled.
// In case the project requieres BasicAuth it will redirect the user to the BasicAuth
// credentials URL. Otherwise, it starts a new crawler and redirects to the live crawling page.
func (h *crawlHandler) startHandler(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	user, ok := h.CookieSession.GetUser(r.Context())
	if !ok {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
	}

	p, err := h.ProjectService.FindProject(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if p.BasicAuth {
		http.Redirect(w, r, "/crawl/auth?id="+strconv.Itoa(pid), http.StatusSeeOther)
		return
	}

	err = h.CrawlerService.StartCrawler(p, models.BasicAuth{})
	if err != nil {
		log.Printf("start crawler for %s error: %v\n", p.URL, err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/crawl/live?pid="+strconv.Itoa(pid), http.StatusSeeOther)
}

// stopHandler handles the crawler stopping.
// It expects a query paramater "pid" containing the id of the project to be stopped.
// Aftar making sure the user owns the project it is stopped.
// In case the request is made via ajax with the X-Requested-With header it will return
// a json response, otherwise it will redirect the user back to the live crawl page.
func (h *crawlHandler) stopHandler(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	user, ok := h.CookieSession.GetUser(r.Context())
	if !ok {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
	}

	p, err := h.ProjectService.FindProject(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	go h.CrawlerService.StopCrawler(p)

	if r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
		data := struct{ Crawling bool }{Crawling: false}
		w.Header().Set("Content-Type", "hlication/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(data)
		return
	}

	http.Redirect(w, r, "/crawl/live?pid="+strconv.Itoa(pid), http.StatusSeeOther)
}

// handleCrawlAuth handles the crawling of a project with BasicAuth.
// It expects a query parameter "pid" containing the project id to be crawled.
// A form will be presented to the user to input the BasicAuth credentials.
// This handler handles the GET request.
func (h *crawlHandler) authGetHandler(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	user, ok := h.CookieSession.GetUser(r.Context())
	if !ok {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
	}

	p, err := h.ProjectService.FindProject(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	pageView := &PageView{
		PageTitle: "CRAWL_AUTH_VIEW_PAGE_TITLE",
		Data:      struct{ Project models.Project }{Project: p},
	}

	h.Renderer.RenderTemplate(w, "crawl_auth", pageView)
}

// Handle the BasicAuth form. Once it is submitted a crawler with BasicAuth is started.
// It processes the auth form data and starts the crawler.
// This handler handles the POST request and redirects to the live crawl page.
func (h *crawlHandler) authPostHandler(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	user, ok := h.CookieSession.GetUser(r.Context())
	if !ok {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
	}

	p, err := h.ProjectService.FindProject(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	err = r.ParseForm()
	if err != nil {
		http.Redirect(w, r, "/crawl/auth", http.StatusSeeOther)
		return
	}

	basicAuth := models.BasicAuth{
		AuthUser: r.FormValue("username"),
		AuthPass: r.FormValue("password"),
	}

	err = h.CrawlerService.StartCrawler(p, basicAuth)
	if err != nil {
		log.Printf("start basic auth crawler for %s error: %v\n", p.URL, err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/crawl/live?pid="+strconv.Itoa(pid), http.StatusSeeOther)
}

// liveCrawlHandler handles the request for the live crawling of a project.
// It expects a query parameter "pid" containing the project id to be crawled.
// This handler renders a page that will connect via websockets to display the progress
// of the crawl.
func (h *crawlHandler) liveCrawlHandler(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	user, ok := h.CookieSession.GetUser(r.Context())
	if !ok {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
	}

	pv, err := h.ProjectViewService.GetProjectView(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if !pv.Crawl.Crawling {
		http.Redirect(w, r, "/dashboard?pid="+strconv.Itoa(pid), http.StatusSeeOther)
		return
	}

	configURL, err := url.Parse(h.Config.HTTPServer.URL)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	v := &PageView{
		Data: struct {
			Project models.Project
			Secure  bool
		}{
			Project: pv.Project,
			Secure:  configURL.Scheme == "https",
		},
		User:      *user,
		PageTitle: "CRAWL_LIVE_PAGE_TITLE",
	}

	h.Renderer.RenderTemplate(w, "crawl_live", v)
}

// wsHandler handles the live crawling of a project using websockets.
// It expects a query parameter "pid" containing the project id.
// It upgrades the connection to websockets and sends the crawler messages through it.
func (h *crawlHandler) wsHandler(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, ok := h.CookieSession.GetUser(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	p, err := h.ProjectService.FindProject(pid, user.Id)
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
			return origin == h.Config.HTTPServer.URL
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("crawlWS upgrader error: %v", err)
		return
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error { conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	connLock := &sync.RWMutex{}

	// Subscribe to the pubsub broker to keep track of the crawl's progress.
	subscriber := h.PubSubBroker.NewSubscriber(fmt.Sprintf("crawl-%d", p.Id), func(i *models.Message) error {
		pubsubMessage := i
		wsMessage := struct {
			Name string
			Data interface{}
		}{
			Name: pubsubMessage.Name,
		}

		if pubsubMessage.Name == "PageReport" {
			msg := pubsubMessage.Data.(*models.PageReportMessage)
			wsMessage.Data = struct {
				StatusCode int
				URL        string
				Crawled    int
				Discovered int
				Crawling   bool
			}{
				StatusCode: msg.StatusCode,
				URL:        msg.URL,
				Crawled:    msg.Crawled,
				Discovered: msg.Discovered,
				Crawling:   msg.Crawling,
			}
		}

		if pubsubMessage.Name == "CrawlEnd" {
			msg := pubsubMessage.Data.(int)
			wsMessage.Data = msg
		}

		connLock.Lock()
		conn.SetWriteDeadline(time.Now().Add(writeWait))
		err := conn.WriteJSON(wsMessage)
		connLock.Unlock()

		return err
	})
	defer h.PubSubBroker.Unsubscribe(subscriber)

	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()

		for range ticker.C {
			connLock.Lock()
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
			connLock.Unlock()
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
