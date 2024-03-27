package routes

import (
	"fmt"
	"log"
	"net/http"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/services"
)

// PageView is the data structure used to render the html templates.
type PageView struct {
	PageTitle string
	User      models.User
	Data      interface{}
	Refresh   bool
}

// NewServer sets up the HTTP server routes and starts the HTTP server.
func NewServer(container *services.Container) {
	// Handle static files
	fileServer := http.FileServer(http.Dir("./web/static"))
	http.Handle("/resources/", http.StripPrefix("/resources", fileServer))
	http.Handle("/robots.txt", fileServer)
	http.Handle("/favicon.ico", fileServer)

	// Crawler routes
	crawlHandler := crawlHandler{container}
	http.HandleFunc("/crawl", container.CookieSession.Auth(crawlHandler.handleCrawl))
	http.HandleFunc("/crawl/stop", container.CookieSession.Auth(crawlHandler.handleStopCrawl))
	http.HandleFunc("/crawl/live", container.CookieSession.Auth(crawlHandler.handleCrawlLive))
	http.HandleFunc("/crawl/auth", container.CookieSession.Auth(crawlHandler.handleCrawlAuth))
	http.HandleFunc("/crawl/ws", container.CookieSession.Auth(crawlHandler.handleCrawlWs))

	// Dashboard route
	dashboardHandler := dashboardHandler{container}
	http.HandleFunc("/dashboard", container.CookieSession.Auth(dashboardHandler.handleDashboard))

	// URL explorer route
	explorerHandler := explorerHandler{container}
	http.HandleFunc("/explorer", container.CookieSession.Auth(explorerHandler.handleExplorer))

	// Data export routes
	exportHandler := exportHandler{container}
	http.HandleFunc("/download", container.CookieSession.Auth(exportHandler.handleDownloadCSV))
	http.HandleFunc("/sitemap", container.CookieSession.Auth(exportHandler.handleSitemap))
	http.HandleFunc("/export", container.CookieSession.Auth(exportHandler.handleExport))
	http.HandleFunc("/export/download", container.CookieSession.Auth(exportHandler.handleExportResources))

	// Issues routes
	issueHandler := issueHandler{container}
	http.HandleFunc("/issues", container.CookieSession.Auth(issueHandler.handleIssues))
	http.HandleFunc("/issues/view", container.CookieSession.Auth(issueHandler.handleIssuesView))

	// Project routes
	projectHandler := projectHandler{container}
	http.HandleFunc("/", container.CookieSession.Auth(projectHandler.handleHome))
	http.HandleFunc("/project/add", container.CookieSession.Auth(projectHandler.handleProjectAdd))
	http.HandleFunc("/project/edit", container.CookieSession.Auth(projectHandler.handleProjectEdit))
	http.HandleFunc("/project/delete", container.CookieSession.Auth(projectHandler.handleDeleteProject))

	// Resource route
	resourceHandler := resourceHandler{container}
	http.HandleFunc("/resources", container.CookieSession.Auth(resourceHandler.handleResourcesView))

	// User routes
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
