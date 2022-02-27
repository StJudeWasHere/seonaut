package app

import (
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/mnlg/lenkrr/internal/report"
	"github.com/mnlg/lenkrr/internal/user"
)

type IssuesGroupView struct {
	IssuesGroups    map[string]IssueGroup
	Project         Project
	Crawl           Crawl
	MediaCount      CountList
	StatusCodeCount CountList
	MediaChart      Chart
	StatusChart     Chart
	Critical        int
	Alert           int
	Warning         int
}

type IssuesView struct {
	PageReports  []report.PageReport
	Cid          int
	Eid          string
	Project      Project
	CurrentPage  int
	NextPage     int
	PreviousPage int
	TotalPages   int
}

func (app *App) serveIssues(user *user.User, w http.ResponseWriter, r *http.Request) {
	qcid, ok := r.URL.Query()["cid"]
	if !ok || len(qcid) < 1 {
		log.Println("serveIssues: cid parameter missing")
		return
	}

	cid, err := strconv.Atoi(qcid[0])
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	u, err := app.datastore.findCrawlUserId(cid)
	if err != nil || u.Id != user.Id {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	issueGroups := app.datastore.findIssues(cid)
	crawl := app.datastore.findCrawlById(cid)
	project, err := app.datastore.findProjectById(crawl.ProjectId, user.Id)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	mediaCount := app.datastore.CountByMediaType(cid)
	mediaChart := NewChart(mediaCount)
	statusCount := app.datastore.CountByStatusCode(cid)
	statusChart := NewChart(statusCount)

	ParsedURL, err := url.Parse(project.URL)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	project.Host = ParsedURL.Host

	var critical int
	var alert int
	var warning int

	for _, v := range issueGroups {
		switch v.Priority {
		case Critical:
			critical += v.Count
		case Alert:
			alert += v.Count
		case Warning:
			warning += v.Count
		}
	}

	ig := IssuesGroupView{
		IssuesGroups:    issueGroups,
		Crawl:           crawl,
		Project:         project,
		MediaCount:      mediaCount,
		MediaChart:      mediaChart,
		StatusChart:     statusChart,
		StatusCodeCount: statusCount,
		Critical:        critical,
		Alert:           alert,
		Warning:         warning,
	}

	v := &PageView{
		Data:      ig,
		User:      *user,
		PageTitle: "ISSUES_VIEW",
	}

	app.renderer.renderTemplate(w, "issues", v)
}

func (app *App) serveIssuesView(user *user.User, w http.ResponseWriter, r *http.Request) {
	qeid, ok := r.URL.Query()["eid"]
	if !ok || len(qeid) < 1 {
		log.Println("serveIssuesView: eid parameter missing")
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	eid := qeid[0]

	qcid, ok := r.URL.Query()["cid"]
	if !ok || len(qcid) < 1 {
		log.Println("serveIssuesView: cid parameter missing")
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	cid, err := strconv.Atoi(qcid[0])
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	totalPages := app.datastore.getNumberOfPagesForIssues(cid, eid)

	p := r.URL.Query()["p"]
	page := 1
	if len(p) > 0 {
		page, err = strconv.Atoi(p[0])
		if err != nil {
			log.Println(err)
			page = 1
		}

		if page < 1 || page > totalPages {
			http.Redirect(w, r, "/", http.StatusSeeOther)

			return
		}
	}

	nextPage := 0
	previousPage := 0

	if page < totalPages {
		nextPage = page + 1
	}

	if page > 1 {
		previousPage = page - 1
	}

	u, err := app.datastore.findCrawlUserId(cid)
	if err != nil || u.Id != user.Id {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	crawl := app.datastore.findCrawlById(cid)
	project, err := app.datastore.findProjectById(crawl.ProjectId, user.Id)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	ParsedURL, err := url.Parse(project.URL)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	project.Host = ParsedURL.Host

	issues := app.datastore.findPageReportIssues(cid, page-1, eid)

	view := IssuesView{
		Cid:          cid,
		Eid:          eid,
		PageReports:  issues,
		Project:      project,
		CurrentPage:  page,
		NextPage:     nextPage,
		PreviousPage: previousPage,
		TotalPages:   totalPages,
	}

	v := &PageView{
		Data:      view,
		User:      *user,
		PageTitle: "ISSUES_DETAIL",
	}

	app.renderer.renderTemplate(w, "issues_view", v)
}
