package http

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/stjudewashere/seonaut/internal/projectview"
	"github.com/stjudewashere/seonaut/internal/report"
)

// handleResourcesView handles the HTTP request for the resources view page.
//
// It expects the following query parameters:
// - "pid" containing the project ID.
// - "rid" the ID of the resource to be loaded.
// - "eid" the ID of the issue type from wich the user loaded this resource.
// - "ep" the explorer page number from which the user loaded this resource.
// - "t" the tab to be loaded, which defaults to the details tab.
// - "p" the number of page to be loaded, in case the resource page has pagination.
func (app *App) handleResourcesView(w http.ResponseWriter, r *http.Request) {
	user, ok := app.userService.GetUserFromContext(r.Context())
	if ok == false {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)

		return
	}

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
	ep := r.URL.Query().Get("ep")
	if eid == "" && ep == "" {
		log.Println("serveResourcesView: no eid or ep parameter set")
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

	pv, err := app.projectViewService.GetProjectView(pid, user.Id)
	if err != nil {
		log.Printf("serveResourcesView GetProjectView: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	data := &struct {
		PageReportView *report.PageReportView
		ProjectView    *projectview.ProjectView
		Eid            string
		Ep             string
		Tab            string
	}{
		ProjectView:    pv,
		Eid:            eid,
		Ep:             ep,
		Tab:            tab,
		PageReportView: app.reportService.GetPageReport(rid, pv.Crawl.Id, tab, page),
	}

	pageView := &PageView{
		Data:      data,
		User:      *user,
		PageTitle: "RESOURCES_VIEW_" + strings.ToUpper(tab),
	}

	app.renderer.RenderTemplate(w, "resources", pageView)
}
