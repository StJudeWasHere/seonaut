package main

import (
	"fmt"
	"html/template"
	"net/http"
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

	view := PageReportView{
		PageReports:           FindPageReports(),
		EmptyTitle:            FindPageReportsWithEmptyTitle(),
		ShortTitle:            FindPageReportsWithShortTitle(),
		LongTitle:             FindPageReportsWithLongTitle(),
		DuplicatedTitle:       FindPageReportsWithDuplicatedTitle(),
		EmptyDescription:      FindPageReportsWithEmptyDescription(),
		ShortDescription:      FindPageReportsWithShortDescription(),
		LongDescription:       FindPageReportsWithLongDescription(),
		DuplicatedDescription: FindPageReportsWithDuplicatedDescription(),
		TotalCount:            CountCrawled(),
		MediaCount:            CountByMediaType(),
		StatusCodeCount:       CountByStatusCode(),
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
