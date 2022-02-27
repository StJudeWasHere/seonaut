package app

import (
	"log"
	"net/http"
	"net/url"
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
	qrid, ok := r.URL.Query()["rid"]
	if !ok || len(qrid) < 1 {
		log.Println("serveResourcesView: rid paramenter missing")
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	rid, err := strconv.Atoi(qrid[0])
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	qcid, ok := r.URL.Query()["cid"]
	if !ok || len(qcid) < 1 {
		log.Println("serveResourcesView: cid parameter missing")
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	cid, err := strconv.Atoi(qcid[0])
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	tabs := r.URL.Query()["t"]
	var tab string
	if len(tabs) == 0 {
		tab = "details"
	} else {
		tab = tabs[0]
	}

	qeid, ok := r.URL.Query()["eid"]
	if !ok || len(qeid) < 1 {
		log.Println("serveResourcesView: eid parameter missing")
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	eid := qeid[0]

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
