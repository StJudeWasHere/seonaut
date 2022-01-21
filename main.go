package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"
)

const (
	port = 9000
	host = "127.0.0.1"
)

func main() {
	crawl := flag.String("crawl", "", "Site to crawl")
	flag.Parse()

	if *crawl != "" {
		fmt.Printf("Crawling %s...\n", string(*crawl))
		start := time.Now()
		startCrawler(Project{URL: string(*crawl)})
		fmt.Println(time.Since(start))
	}

	http.HandleFunc("/", serveHome)
	http.HandleFunc("/new-project", serveProjectAdd)
	http.HandleFunc("/crawl", serveCrawl)
	http.HandleFunc("/issues", serveIssues)
	http.HandleFunc("/issues/view", serveIssuesView)
	http.HandleFunc("/resources", serveResourcesView)
	http.HandleFunc("/signup", serveSignup)
	http.HandleFunc("/signin", serveSignin)

	fmt.Printf("Starting server at %s on port %d...\n", host, port)

	err := http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil)
	if err != nil {
		fmt.Println(err)
	}
}
