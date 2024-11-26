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
	http.Handle("GET /resources/", http.StripPrefix("/resources", fileServer))
	http.Handle("GET /robots.txt", fileServer)
	http.Handle("GET /favicon.ico", fileServer)

	// Crawler routes
	crawlHandler := crawlHandler{container}
	http.HandleFunc("GET /crawl/start", container.CookieSession.Auth(crawlHandler.startHandler))
	http.HandleFunc("GET /crawl/stop", container.CookieSession.Auth(crawlHandler.stopHandler))
	http.HandleFunc("GET /crawl/live", container.CookieSession.Auth(crawlHandler.liveCrawlHandler))
	http.HandleFunc("GET /crawl/auth", container.CookieSession.Auth(crawlHandler.authGetHandler))
	http.HandleFunc("POST /crawl/auth", container.CookieSession.Auth(crawlHandler.authPostHandler))
	http.HandleFunc("GET /crawl/ws", container.CookieSession.Auth(crawlHandler.wsHandler))

	// Dashboard route
	dashboardHandler := dashboardHandler{container}
	http.HandleFunc("GET /dashboard", container.CookieSession.Auth(dashboardHandler.indexHandler))

	// URL explorer route
	explorerHandler := explorerHandler{container}
	http.HandleFunc("GET /explorer", container.CookieSession.Auth(explorerHandler.indexHandler))

	// Data export routes
	exportHandler := exportHandler{container}
	http.HandleFunc("GET /export", container.CookieSession.Auth(exportHandler.indexHandler))
	http.HandleFunc("GET /export/csv", container.CookieSession.Auth(exportHandler.csvHandler))
	http.HandleFunc("GET /export/sitemap", container.CookieSession.Auth(exportHandler.sitemapHandler))
	http.HandleFunc("GET /export/resources", container.CookieSession.Auth(exportHandler.resourcesHandler))
	http.HandleFunc("GET /export/wazc", container.CookieSession.Auth(exportHandler.waczHandler))

	// Issues routes
	issueHandler := issueHandler{container}
	http.HandleFunc("GET /issues", container.CookieSession.Auth(issueHandler.indexHandler))
	http.HandleFunc("GET /issues/view", container.CookieSession.Auth(issueHandler.viewHandler))

	// Project routes
	projectHandler := projectHandler{container}
	http.HandleFunc("GET /", container.CookieSession.Auth(projectHandler.indexHandler))
	http.HandleFunc("GET /project/add", container.CookieSession.Auth(projectHandler.addGetHandler))
	http.HandleFunc("POST /project/add", container.CookieSession.Auth(projectHandler.addPostHandler))
	http.HandleFunc("GET /project/edit", container.CookieSession.Auth(projectHandler.editGetHandler))
	http.HandleFunc("POST /project/edit", container.CookieSession.Auth(projectHandler.editPostHandler))
	http.HandleFunc("GET /project/delete", container.CookieSession.Auth(projectHandler.deleteHandler))

	// Resource route
	resourceHandler := resourceHandler{container}
	http.HandleFunc("GET /resources", container.CookieSession.Auth(resourceHandler.indexHandler))
	http.HandleFunc("GET /archive", container.CookieSession.Auth(resourceHandler.archiveHandler))

	// User routes
	userHandler := userHandler{container}
	http.HandleFunc("GET /signup", userHandler.signupGetHandler)
	http.HandleFunc("POST /signup", userHandler.signupPostHandler)
	http.HandleFunc("GET /signin", userHandler.signinGetHandler)
	http.HandleFunc("POST /signin", userHandler.signinPostHandler)
	http.HandleFunc("GET /account", container.CookieSession.Auth(userHandler.editGetHandler))
	http.HandleFunc("POST /account", container.CookieSession.Auth(userHandler.editPostHandler))
	http.HandleFunc("GET /account/delete", container.CookieSession.Auth((userHandler.deleteGetHandler)))
	http.HandleFunc("POST /account/delete", container.CookieSession.Auth((userHandler.deletePostHandler)))
	http.HandleFunc("GET /signout", container.CookieSession.Auth(userHandler.signoutHandler))

	fmt.Printf("Starting server at %s on port %d...\n", container.Config.HTTPServer.Server, container.Config.HTTPServer.Port)
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", container.Config.HTTPServer.Server, container.Config.HTTPServer.Port), nil)
	if err != nil {
		log.Fatalf("error starting server: %v", err)
	}
}
