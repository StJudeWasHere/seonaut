package main

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"time"
)

const (
	port = 9000
	host = "127.0.0.1"
)

type PageReportView struct {
	PageReports           []PageReport
	EmptyTitle            []PageReport
	ShortTitle            []PageReport
	LongTitle             []PageReport
	DuplicatedTitle       []PageReport
	EmptyDescription      []PageReport
	ShortDescription      []PageReport
	LongDescription       []PageReport
	DuplicatedDescription []PageReport
	TotalCount            int
	MediaCount            map[string]int
	StatusCodeCount       map[int]int
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	//	http.ServeFile(w, r, "home.html")

	pageReports := FindPageReports()
	emptyTitle := FindPageReportsWithEmptyTitle()
	shortTitle := FindPageReportsWithShortTitle()
	longTitle := FindPageReportsWithLongTitle()
	duplicatedTitle := FindPageReportsWithDuplicatedTitle()
	emptyDescription := FindPageReportsWithEmptyDescription()
	shortDescription := FindPageReportsWithShortDescription()
	longDescription := FindPageReportsWithLongDescription()
	duplicatedDescription := FindPageReportsWithDuplicatedDescription()
	totalCount := CountCrawled()
	mediaCount := CountByMediaType()
	statusCodeCount := CountByStatusCode()

	view := PageReportView{
		PageReports:           pageReports,
		EmptyTitle:            emptyTitle,
		ShortTitle:            shortTitle,
		LongTitle:             longTitle,
		DuplicatedTitle:       duplicatedTitle,
		EmptyDescription:      emptyDescription,
		ShortDescription:      shortDescription,
		LongDescription:       longDescription,
		DuplicatedDescription: duplicatedDescription,
		TotalCount:            totalCount,
		MediaCount:            mediaCount,
		StatusCodeCount:       statusCodeCount,
	}

	var templates = template.Must(template.ParseFiles("home.html"))
	templates.ExecuteTemplate(w, "home.html", view)
}

func main() {
	crawl := flag.String("crawl", "", "Site to crawl")
	flag.Parse()

	if *crawl != "" {
		fmt.Printf("Crawling %s...\n", string(*crawl))
		start := time.Now()
		startCrawler(string(*crawl))
		fmt.Println(time.Since(start))
	}

	http.HandleFunc("/", serveHome)

	fmt.Printf("Starting at %s on port %d...\n", host, port)

	err := http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil)
	if err != nil {
		fmt.Println(err)
	}
}

func startCrawler(s string) {
	var crawled int

	pageReport := make(chan PageReport)
	c := &Crawler{}

	u, err := url.Parse(s)
	if err != nil {
		fmt.Println(err)
		return
	}

	go c.Crawl(u, pageReport)

	for r := range pageReport {
		crawled++
		// handlePageReport(r)
		savePageReport(&r)
	}

	fmt.Printf("%d pages crawled.\n", crawled)
}
