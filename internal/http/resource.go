package http

import (
	"log"
	"net/http"
	"strconv"

	"github.com/mnlg/lenkrr/internal/project"
	"github.com/mnlg/lenkrr/internal/report"
	"github.com/mnlg/lenkrr/internal/user"
)

type ResourcesView struct {
	PageReportView *report.PageReportView
	ProjectView    *project.ProjectView
	Eid            string
	Tab            string
}

func (app *App) serveResourcesView(user *user.User, w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		log.Printf("serveResourcesView pid: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	rid, err := strconv.Atoi(r.URL.Query().Get("rid"))
	if err != nil {
		log.Printf("serveResourcesView rid: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	eid := r.URL.Query().Get("eid")
	if eid == "" {
		log.Println("serveResourcesView: eid parameter missing")
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	tab := r.URL.Query().Get("t")
	if tab == "" {
		tab = "details"
	}

	pv, err := app.projectService.GetProjectView(pid, user.Id)
	if err != nil {
		log.Printf("serveResourcesView GetProjectView: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	rv := ResourcesView{
		ProjectView:    pv,
		Eid:            eid,
		Tab:            tab,
		PageReportView: app.reportService.GetPageReport(rid, pv.Crawl.Id, tab),
	}

	v := &PageView{
		Data:      rv,
		User:      *user,
		PageTitle: "RESOURCES_VIEW",
	}

	app.renderer.renderTemplate(w, "resources", v)
}
