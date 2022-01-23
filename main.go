package main

import (
	"fmt"
	"log"
	"net/http"
)

const (
	port = 9000
	host = "127.0.0.1"
)

func main() {
	http.HandleFunc("/", requireAuth(serveHome))
	http.HandleFunc("/new-project", requireAuth(serveProjectAdd))
	http.HandleFunc("/crawl", requireAuth(serveCrawl))
	http.HandleFunc("/issues", requireAuth(serveIssues))
	http.HandleFunc("/issues/view", requireAuth(serveIssuesView))
	http.HandleFunc("/resources", requireAuth(serveResourcesView))
	http.HandleFunc("/signup", serveSignup)
	http.HandleFunc("/signin", serveSignin)

	fmt.Printf("Starting server at %s on port %d...\n", host, port)
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil)
	if err != nil {
		log.Println(err)
	}
}
