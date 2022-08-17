package http

import (
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/stjudewashere/seonaut/internal/project"
	"github.com/stjudewashere/seonaut/internal/projectview"
)

func (app *App) serveHome(w http.ResponseWriter, r *http.Request) {
	user, ok := app.userService.GetUserFromContext(r.Context())
	if ok == false {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
	}

	views := app.projectViewService.GetProjectViews(user.Id)

	var refresh bool
	for _, v := range views {
		if v.Crawl.Id > 0 && (v.Crawl.IssuesEnd.Valid == false || v.Project.Deleting) {
			refresh = true
		}
	}

	v := &PageView{
		Data: struct {
			Projects []projectview.ProjectView
		}{Projects: views},
		User:      *user,
		PageTitle: "PROJECTS_VIEW",
		Refresh:   refresh,
	}

	app.renderer.RenderTemplate(w, "home", v)
}

// Manage the form for adding new projects
func (app *App) serveProjectAdd(w http.ResponseWriter, r *http.Request) {
	user, ok := app.userService.GetUserFromContext(r.Context())
	if ok == false {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Printf("serveProjectAdd ParseForm: %v\n", err)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		u := r.FormValue("url")

		ignoreRobotsTxt, err := strconv.ParseBool(r.FormValue("ignore_robotstxt"))
		if err != nil {
			ignoreRobotsTxt = false
		}

		followNofollow, err := strconv.ParseBool(r.FormValue("follow_nofollow"))
		if err != nil {
			followNofollow = false
		}

		includeNoindex, err := strconv.ParseBool(r.FormValue("include_noindex"))
		if err != nil {
			includeNoindex = false
		}

		crawlSitemap, err := strconv.ParseBool(r.FormValue("crawl_sitemap"))
		if err != nil {
			crawlSitemap = false
		}

		allowSubdomains, err := strconv.ParseBool(r.FormValue("allow_subdomains"))
		if err != nil {
			allowSubdomains = false
		}

		parsedURL, err := url.ParseRequestURI(strings.TrimSpace(u))
		if err != nil {
			v := &PageView{
				User:      *user,
				PageTitle: "ADD_PROJECT",
				Data:      struct{ Error bool }{Error: true},
			}

			app.renderer.RenderTemplate(w, "project_add", v)
			return
		}

		project := &project.Project{
			URL:             parsedURL.String(),
			IgnoreRobotsTxt: ignoreRobotsTxt,
			FollowNofollow:  followNofollow,
			IncludeNoindex:  includeNoindex,
			CrawlSitemap:    crawlSitemap,
			AllowSubdomains: allowSubdomains,
		}

		err = app.projectService.SaveProject(project, user.Id)
		if err != nil {
			v := &PageView{
				User:      *user,
				PageTitle: "ADD_PROJECT",
				Data:      struct{ Error bool }{Error: true},
			}

			app.renderer.RenderTemplate(w, "project_add", v)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	v := &PageView{
		User:      *user,
		PageTitle: "ADD_PROJECT",
		Data:      struct{ Error bool }{Error: false},
	}

	app.renderer.RenderTemplate(w, "project_add", v)
}

// Delete a project
func (app *App) serveDeleteProject(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	user, ok := app.userService.GetUserFromContext(r.Context())
	if ok == false {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
	}

	p, err := app.projectService.FindProject(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	app.projectService.DeleteProject(&p)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Edit project
func (app *App) serveProjectEdit(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	user, ok := app.userService.GetUserFromContext(r.Context())
	if ok == false {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
	}

	p, err := app.projectService.FindProject(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := struct {
		Project project.Project
		Error   bool
	}{Project: p}

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Printf("serveProjectEdit ParseForm: %v\n", err)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		p.IgnoreRobotsTxt, err = strconv.ParseBool(r.FormValue("ignore_robotstxt"))
		if err != nil {
			p.IgnoreRobotsTxt = false
		}

		p.FollowNofollow, err = strconv.ParseBool(r.FormValue("follow_nofollow"))
		if err != nil {
			p.FollowNofollow = false
		}

		p.IncludeNoindex, err = strconv.ParseBool(r.FormValue("include_noindex"))
		if err != nil {
			p.IncludeNoindex = false
		}

		p.CrawlSitemap, err = strconv.ParseBool(r.FormValue("crawl_sitemap"))
		if err != nil {
			p.CrawlSitemap = false
		}

		p.AllowSubdomains, err = strconv.ParseBool(r.FormValue("allow_subdomains"))
		if err != nil {
			p.AllowSubdomains = false
		}

		err = app.projectService.UpdateProject(&p)
		if err != nil {
			data.Error = true
			v := &PageView{
				User:      *user,
				PageTitle: "EDIT_PROJECT",
				Data:      data,
			}

			app.renderer.RenderTemplate(w, "project_edit", v)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	v := &PageView{
		User:      *user,
		PageTitle: "EDIT_PROJECT",
		Data:      data,
	}

	app.renderer.RenderTemplate(w, "project_edit", v)
}
