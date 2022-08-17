package http

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/stjudewashere/seonaut/internal/projectview"
	"github.com/stjudewashere/seonaut/internal/report"
)

type ResourcesView struct {
	PageReportView *report.PageReportView
	ProjectView    *projectview.ProjectView
	Eid            string
	Tab            string
}

func (app *App) serveResourcesView(w http.ResponseWriter, r *http.Request) {
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

	page, err := strconv.Atoi(r.URL.Query().Get("p"))
	if err != nil {
		page = 1
	}

	user, ok := app.userService.GetUserFromContext(r.Context())
	if ok == false {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
	}

	pv, err := app.projectViewService.GetProjectView(pid, user.Id)
	if err != nil {
		log.Printf("serveResourcesView GetProjectView: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	rv := ResourcesView{
		ProjectView:    pv,
		Eid:            eid,
		Tab:            tab,
		PageReportView: app.reportService.GetPageReport(rid, pv.Crawl.Id, tab, page),
	}

	v := &PageView{
		Data:      rv,
		User:      *user,
		PageTitle: "RESOURCES_VIEW_" + strings.ToUpper(tab),
	}

	app.renderer.RenderTemplate(w, "resources", v)
}
