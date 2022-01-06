package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	port = 9000
	host = "127.0.0.1"
)

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	http.ServeFile(w, r, "home.html")
}

func main() {
	/*
		REPORT:
		create sitemap
		create image sitemap
		detect redirect chains
		detect redirect loops
		detect duplicated or empty description/title
	*/

	/*
		http.HandleFunc("/", serveHome)

		fmt.Printf("Starting at %s on port %d...\n", host, port)

		err := http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil)
		if err != nil {
			fmt.Println(err)
		}
	*/
	start := time.Now()

	startCrawler()

	fmt.Println(time.Since(start))
}

func startCrawler() {
	var crawled int
	a := os.Args[1]

	pageReport := make(chan PageReport)
	c := &Crawler{}

	u, err := url.Parse(a)
	if err != nil {
		fmt.Println(err)
		return
	}

	go c.Crawl(u, pageReport)

	for r := range pageReport {
		crawled++
		handlePageReport(r)
	}

	writer.Flush()

	fmt.Printf("%d pages crawled.\n", crawled)

}
