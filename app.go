package main

import (
	"fmt"
	"log"

	"net/http"

	"github.com/gorilla/sessions"
)

type App struct {
	config    *Config
	datastore *datastore
	cookie    *sessions.CookieStore
}

func NewApp(configPath string) *App {
	var err error
	var app App

	app.config, err = loadConfig(configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v\n", err)
	}

	app.datastore, err = NewDataStore(
		app.config.DbUser,
		app.config.DbPass,
		app.config.DbServer,
		app.config.DbPort,
		app.config.DbName,
	)

	if err != nil {
		log.Fatalf("Error creating new datastore: %v\n", err)
	}

	app.cookie = sessions.NewCookieStore([]byte("SESSION_ID"))

	return &app
}

func (app *App) Run() {
	fileServer := http.FileServer(http.Dir("./static"))
	http.Handle("/resources/", http.StripPrefix("/resources", fileServer))
	http.Handle("/robots.txt", fileServer)
	http.Handle("/favicon.ico", fileServer)

	http.HandleFunc("/", app.requireAuth(app.serveHome))
	http.HandleFunc("/new-project", app.requireAuth(app.serveProjectAdd))
	http.HandleFunc("/crawl", app.requireAuth(app.serveCrawl))
	http.HandleFunc("/issues", app.requireAuth(app.serveIssues))
	http.HandleFunc("/issues/view", app.requireAuth(app.serveIssuesView))
	http.HandleFunc("/download", app.requireAuth(app.serveDownloadAll))
	http.HandleFunc("/resources", app.requireAuth(app.serveResourcesView))
	http.HandleFunc("/signup", app.serveSignup)
	http.HandleFunc("/signin", app.serveSignin)
	http.HandleFunc("/signout", app.requireAuth(app.serveSignout))

	fmt.Printf("Starting server at %s on port %d...\n", app.config.Server, app.config.ServerPort)
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", app.config.Server, app.config.ServerPort), nil)
	if err != nil {
		log.Fatal(err)
	}
}
