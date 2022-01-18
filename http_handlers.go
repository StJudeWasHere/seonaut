package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"
)

type PageReportView struct {
	Projects              []Project
	Crawl                 Crawl
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

type Crawl struct {
	Id    int
	URL   string
	Start time.Time
	End   time.Time
}

type Project struct {
	Id      int
	URL     string
	Created time.Time
}

func (c Crawl) TotalTime() time.Duration {
	return c.End.Sub(c.Start)
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

	projects := findProjects()

	crawl := getLastCrawl()
	cid := crawl.Id

	view := PageReportView{
		//		PageReports:           FindPageReports(),
		Projects:              projects,
		Crawl:                 crawl,
		EmptyTitle:            FindPageReportsWithEmptyTitle(cid),
		ShortTitle:            FindPageReportsWithShortTitle(cid),
		LongTitle:             FindPageReportsWithLongTitle(cid),
		DuplicatedTitle:       FindPageReportsWithDuplicatedTitle(cid),
		EmptyDescription:      FindPageReportsWithEmptyDescription(cid),
		ShortDescription:      FindPageReportsWithShortDescription(cid),
		LongDescription:       FindPageReportsWithLongDescription(cid),
		DuplicatedDescription: FindPageReportsWithDuplicatedDescription(cid),
		TotalCount:            CountCrawled(cid),
		MediaCount:            CountByMediaType(cid),
		StatusCodeCount:       CountByStatusCode(cid),
	}

	var templates = template.Must(template.ParseFiles(
		"templates/home.html", "templates/head.html", "templates/footer.html", "templates/list.html",
		"templates/url_list.html", "templates/pagereport.html",
	))

	err := templates.ExecuteTemplate(w, "home.html", view)
	if err != nil {
		fmt.Println(err)
	}
}

func serveProjectAdd(w http.ResponseWriter, r *http.Request) {
	var url string

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			fmt.Println(err)
		}
		url = r.FormValue("url")
		saveProject(url)

	}

	var templates = template.Must(template.ParseFiles(
		"templates/project_add.html", "templates/head.html", "templates/footer.html",
	))

	err := templates.ExecuteTemplate(w, "project_add.html", struct{ URL string }{URL: url})
	if err != nil {
		fmt.Println(err)
	}
}

func serveCrawl(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query()["pid"][0])
	fmt.Println(pid)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/", 303)
	}

	p := findProjectById(pid)
	fmt.Println(p)
	fmt.Printf("Crawling %s...\n", p.URL)
	go func() {
		start := time.Now()
		startCrawler(p)
		fmt.Println(time.Since(start))
	}()

	http.Redirect(w, r, "/", 303)
}
