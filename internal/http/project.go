package http

import (
	"log"
	"net/http"
	"strconv"

	"github.com/mnlg/lenkrr/internal/project"

	"github.com/mnlg/lenkrr/internal/user"
)

func (app *App) serveHome(user *user.User, w http.ResponseWriter, r *http.Request) {
	views := app.projectService.GetProjectViews(user.Id)

	var refresh bool
	for _, v := range views {
		if v.Crawl.IssuesEnd.Valid == false {
			refresh = true
		}
	}

	v := &PageView{
		Data: struct {
			Projects    []project.ProjectView
			MaxProjects int
		}{Projects: views, MaxProjects: user.GetMaxAllowedProjects()},
		User:      *user,
		PageTitle: "PROJECTS_VIEW",
		Refresh:   refresh,
	}

	app.renderer.renderTemplate(w, "home", v)
}

func (app *App) serveProjectAdd(user *user.User, w http.ResponseWriter, r *http.Request) {
	projects := app.projectService.GetProjects(user.Id)
	if len(projects) >= user.GetMaxAllowedProjects() {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Println("serveProjectAdd ParseForm: %v\n", err)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		url := r.FormValue("url")

		ignoreRobotsTxt, err := strconv.ParseBool(r.FormValue("ignore_robotstxt"))
		if err != nil {
			ignoreRobotsTxt = false
		}

		app.projectService.SaveProject(url, ignoreRobotsTxt, user.Id)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	v := &PageView{
		User:      *user,
		PageTitle: "ADD_PROJECT",
	}

	app.renderer.renderTemplate(w, "project_add", v)
}
