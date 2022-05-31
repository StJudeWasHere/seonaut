package http

import (
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/stjudewashere/seonaut/internal/helper"
	"github.com/stjudewashere/seonaut/internal/project"
	"github.com/stjudewashere/seonaut/internal/projectview"
	"github.com/stjudewashere/seonaut/internal/user"
)

func (app *App) serveHome(w http.ResponseWriter, r *http.Request) {
	c := r.Context().Value("user")
	user, ok := c.(*user.User)
	if ok == false {
		http.Redirect(w, r, "/signout", http.StatusSeeOther)
		return
	}

	views := app.projectViewService.GetProjectViews(user.Id)

	var refresh bool
	for _, v := range views {
		if v.Crawl.IssuesEnd.Valid == false {
			refresh = true
		}
	}

	v := &helper.PageView{
		Data: struct {
			Projects []projectview.ProjectView
		}{Projects: views},
		User:      *user,
		PageTitle: "PROJECTS_VIEW",
		Refresh:   refresh,
	}

	app.renderer.RenderTemplate(w, "home", v)
}

func (app *App) serveProjectAdd(w http.ResponseWriter, r *http.Request) {
	c := r.Context().Value("user")
	user, ok := c.(*user.User)
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
			v := &helper.PageView{
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
			v := &helper.PageView{
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

	v := &helper.PageView{
		User:      *user,
		PageTitle: "ADD_PROJECT",
		Data:      struct{ Error bool }{Error: false},
	}

	app.renderer.RenderTemplate(w, "project_add", v)
}
