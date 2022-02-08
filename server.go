package main

import (
	"fmt"
	"log"
	"net/http"
)

func initServer() {
	fileServer := http.FileServer(http.Dir("./static"))
	http.Handle("/resources/", http.StripPrefix("/resources", fileServer))
	http.Handle("/robots.txt", fileServer)
	http.Handle("/favicon.ico", fileServer)

	http.HandleFunc("/", requireAuth(serveHome))
	http.HandleFunc("/new-project", requireAuth(serveProjectAdd))
	http.HandleFunc("/crawl", requireAuth(serveCrawl))
	http.HandleFunc("/issues", requireAuth(serveIssues))
	http.HandleFunc("/issues/view", requireAuth(serveIssuesView))
	http.HandleFunc("/download", requireAuth(serveDownloadAll))
	http.HandleFunc("/resources", requireAuth(serveResourcesView))
	http.HandleFunc("/signup", serveSignup)
	http.HandleFunc("/signin", serveSignin)
	http.HandleFunc("/signout", requireAuth(serveSignout))

	fmt.Printf("Starting server at %s on port %d...\n", config.Server, config.ServerPort)
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", config.Server, config.ServerPort), nil)
	if err != nil {
		log.Fatal(err)
	}
}
