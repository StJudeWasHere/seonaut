package http

import (
	"fmt"
	"log"
	"net/http"

	"github.com/stjudewashere/seonaut/internal/container"
	"github.com/stjudewashere/seonaut/internal/models"
)

// PageView is the data structure used to render the html templates.
type PageView struct {
	PageTitle string
	User      models.User
	Data      interface{}
	Refresh   bool
}

// NewApp initializes the template renderer and the session cookie.
// Returns a new HTTP application server.
func NewServer(container *container.Container) {
	// Static
	fileServer := http.FileServer(http.Dir("./web/static"))
	http.Handle("/resources/", http.StripPrefix("/resources", fileServer))
	http.Handle("/robots.txt", fileServer)
	http.Handle("/favicon.ico", fileServer)

	// App
	crawlHandler := crawlHandler{container}
	http.HandleFunc("/crawl", container.CookieSession.Auth(crawlHandler.handleCrawl))
	http.HandleFunc("/crawl/stop", container.CookieSession.Auth(crawlHandler.handleStopCrawl))
	http.HandleFunc("/crawl/live", container.CookieSession.Auth(crawlHandler.handleCrawlLive))
	http.HandleFunc("/crawl/auth", container.CookieSession.Auth(crawlHandler.handleCrawlAuth))
	http.HandleFunc("/crawl/ws", container.CookieSession.Auth(crawlHandler.handleCrawlWs))

	dashboardHandler := dashboardHandler{container}
	http.HandleFunc("/dashboard", container.CookieSession.Auth(dashboardHandler.handleDashboard))

	explorerHandler := explorerHandler{container}
	http.HandleFunc("/explorer", container.CookieSession.Auth(explorerHandler.handleExplorer))

	exportHandler := exportHandler{container}
	http.HandleFunc("/download", container.CookieSession.Auth(exportHandler.handleDownloadCSV))
	http.HandleFunc("/sitemap", container.CookieSession.Auth(exportHandler.handleSitemap))
	http.HandleFunc("/export", container.CookieSession.Auth(exportHandler.handleExport))
	http.HandleFunc("/export/download", container.CookieSession.Auth(exportHandler.handleExportResources))

	issueHandler := issueHandler{container}
	http.HandleFunc("/issues", container.CookieSession.Auth(issueHandler.handleIssues))
	http.HandleFunc("/issues/view", container.CookieSession.Auth(issueHandler.handleIssuesView))

	projectHandler := projectHandler{container}
	http.HandleFunc("/", container.CookieSession.Auth(projectHandler.handleHome))
	http.HandleFunc("/project/add", container.CookieSession.Auth(projectHandler.handleProjectAdd))
	http.HandleFunc("/project/edit", container.CookieSession.Auth(projectHandler.handleProjectEdit))
	http.HandleFunc("/project/delete", container.CookieSession.Auth(projectHandler.handleDeleteProject))

	resourceHandler := resourceHandler{container}
	http.HandleFunc("/resources", container.CookieSession.Auth(resourceHandler.handleResourcesView))

	userHandle := userHandler{container}
	http.HandleFunc("/signup", userHandle.handleSignup)
	http.HandleFunc("/signin", userHandle.handleSignin)
	http.HandleFunc("/signout", container.CookieSession.Auth(userHandle.handleSignout))
	http.HandleFunc("/account", container.CookieSession.Auth(userHandle.handleAccount))
	http.HandleFunc("/account/delete", container.CookieSession.Auth((userHandle.handleDeleteUser)))

	fmt.Printf("Starting server at %s on port %d...\n", container.Config.HTTPServer.Server, container.Config.HTTPServer.Port)
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", container.Config.HTTPServer.Server, container.Config.HTTPServer.Port), nil)
	if err != nil {
		log.Fatal(err)
	}
}
