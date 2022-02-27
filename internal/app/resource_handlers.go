package app

import (
	"log"
	"net/http"
	"strconv"

	"github.com/mnlg/lenkrr/internal/project"
	"github.com/mnlg/lenkrr/internal/report"
	"github.com/mnlg/lenkrr/internal/user"
)

type ResourcesView struct {
	PageReport report.PageReport
	Cid        int
	Eid        string
	ErrorTypes []string
	InLinks    []report.PageReport
	Redirects  []report.PageReport
	Project    project.Project
	Tab        string
}

func (app *App) serveResourcesView(user *user.User, w http.ResponseWriter, r *http.Request) {
	rid, err := strconv.Atoi(r.URL.Query().Get("rid"))
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	cid, err := strconv.Atoi(r.URL.Query().Get("cid"))
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	tab := r.URL.Query().Get("t")
	if tab == "" {
		tab = "details"
	}

	eid := r.URL.Query().Get("eid")
	if eid == "" {
		log.Println("serveResourcesView: eid parameter missing")
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
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

	pageReport := app.datastore.FindPageReportById(rid)
	errorTypes := app.datastore.findErrorTypesByPage(rid, cid)

	var inLinks []report.PageReport
	if tab == "inlinks" {
		inLinks = app.datastore.FindInLinks(pageReport.URL, cid)
	}

	var redirects []report.PageReport
	if tab == "redirections" {
		redirects = app.datastore.FindPageReportsRedirectingToURL(pageReport.URL, cid)
	}

	rv := ResourcesView{
		PageReport: pageReport,
		Project:    project,
		Cid:        cid,
		Eid:        eid,
		ErrorTypes: errorTypes,
		InLinks:    inLinks,
		Redirects:  redirects,
		Tab:        tab,
	}

	v := &PageView{
		Data:      rv,
		User:      *user,
		PageTitle: "RESOURCES_VIEW",
	}

	app.renderer.renderTemplate(w, "resources", v)
}
