package app

import (
	"log"
	"net/http"
	"strconv"

	"github.com/mnlg/lenkrr/internal/crawler"
	"github.com/mnlg/lenkrr/internal/issue"
	"github.com/mnlg/lenkrr/internal/project"
	"github.com/mnlg/lenkrr/internal/report"
	"github.com/mnlg/lenkrr/internal/user"
)

type IssuesGroupView struct {
	Project     project.Project
	Crawl       crawler.Crawl
	MediaChart  Chart
	StatusChart Chart
	IssueCount  *issue.IssueCount
}

type IssuesView struct {
	PageReports  []report.PageReport
	Cid          int
	Eid          string
	Project      project.Project
	CurrentPage  int
	NextPage     int
	PreviousPage int
	TotalPages   int
}

func (app *App) serveIssues(user *user.User, w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	project, err := app.projectService.FindProject(pid, user.Id)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	crawl := app.datastore.GetLastCrawl(&project)
	issueCount := app.issueService.GetIssuesCount(crawl.Id)

	ig := IssuesGroupView{
		Project:     project,
		Crawl:       crawl,
		MediaChart:  NewChart(issueCount.MediaCount),
		StatusChart: NewChart(issueCount.StatusCount),
		IssueCount:  issueCount,
	}

	v := &PageView{
		Data:      ig,
		User:      *user,
		PageTitle: "ISSUES_VIEW",
	}

	app.renderer.renderTemplate(w, "issues", v)
}

func (app *App) serveIssuesView(user *user.User, w http.ResponseWriter, r *http.Request) {
	eid := r.URL.Query().Get("eid")
	if eid == "" {
		log.Println("serveIssuesView: eid parameter missing")
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	cid, err := strconv.Atoi(r.URL.Query().Get("cid"))
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	totalPages := app.datastore.getNumberOfPagesForIssues(cid, eid)

	page, err := strconv.Atoi(r.URL.Query().Get("p"))
	if err != nil {
		page = 1
	}

	if page < 1 || page > totalPages {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
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
	project, err := app.projectService.FindProject(crawl.ProjectId, user.Id)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

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
