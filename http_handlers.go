package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id       int
	Email    string
	Password string
	Advanced bool
}

type ProjectView struct {
	Project Project
	Crawl   Crawl
}

type IssuesGroupView struct {
	IssuesGroups    map[string]IssueGroup
	Project         Project
	Crawl           Crawl
	MediaCount      CountList
	StatusCodeCount CountList
	MediaChart      Chart
	StatusChart     Chart
}

type IssuesView struct {
	PageReports  []PageReport
	Cid          int
	Eid          string
	Project      Project
	CurrentPage  int
	NextPage     int
	PreviousPage int
	TotalPages   int
}

type ResourcesView struct {
	PageReport PageReport
	Cid        int
	Eid        string
	ErrorTypes []string
	InLinks    []PageReport
	Redirects  []PageReport
}

type Project struct {
	Id              int
	URL             string
	Host            string
	IgnoreRobotsTxt bool
	UseJS           bool
	Created         time.Time
}

func (app *App) serveHome(user *User, w http.ResponseWriter, r *http.Request) {
	var views []ProjectView
	projects := app.datastore.findProjectsByUser(user.Id)

	for _, p := range projects {
		c := app.datastore.getLastCrawl(&p)
		pv := ProjectView{
			Project: p,
			Crawl:   c,
		}
		views = append(views, pv)
	}

	v := &PageView{
		Data:      views,
		User:      *user,
		PageTitle: "PROJECTS_VIEW",
	}

	renderTemplate(w, "home", v)
}

func (app *App) serveProjectAdd(user *User, w http.ResponseWriter, r *http.Request) {
	var url string

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}
		url = r.FormValue("url")
		ignoreRobotsTxt, err := strconv.ParseBool(r.FormValue("ignore_robotstxt"))
		if err != nil {
			ignoreRobotsTxt = false
		}

		useJavascript, err := strconv.ParseBool(r.FormValue("use_javascript"))
		if err != nil {
			useJavascript = false
		}

		if user.Advanced == false {
			useJavascript = false
		}

		app.datastore.saveProject(url, ignoreRobotsTxt, useJavascript, user.Id)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	v := &PageView{
		User:      *user,
		PageTitle: "ADD_PROJECT",
	}

	renderTemplate(w, "project_add", v)
}

func (app *App) serveCrawl(user *User, w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query()["pid"][0])
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	p, err := app.datastore.findProjectById(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	fmt.Printf("Crawling %s...\n", p.URL)
	go func() {
		start := time.Now()
		cid := startCrawler(p, app.config.CrawlerAgent, app.datastore)
		fmt.Println(time.Since(start))
		fmt.Printf("Creating issues for crawl id %d.\n", cid)

		rm := ReportManager{}

		rm.addReporter(app.datastore.Find30xPageReports, Error30x)
		rm.addReporter(app.datastore.Find40xPageReports, Error40x)
		rm.addReporter(app.datastore.Find50xPageReports, Error50x)
		rm.addReporter(app.datastore.FindPageReportsWithDuplicatedTitle, ErrorDuplicatedTitle)
		rm.addReporter(app.datastore.FindPageReportsWithDuplicatedTitle, ErrorDuplicatedDescription)
		rm.addReporter(app.datastore.FindPageReportsWithEmptyTitle, ErrorEmptyTitle)
		rm.addReporter(app.datastore.FindPageReportsWithShortTitle, ErrorShortTitle)
		rm.addReporter(app.datastore.FindPageReportsWithLongTitle, ErrorLongTitle)
		rm.addReporter(app.datastore.FindPageReportsWithEmptyDescription, ErrorEmptyDescription)
		rm.addReporter(app.datastore.FindPageReportsWithShortDescription, ErrorShortDescription)
		rm.addReporter(app.datastore.FindPageReportsWithLongDescription, ErrorLongDescription)
		rm.addReporter(app.datastore.FindPageReportsWithLittleContent, ErrorLittleContent)
		rm.addReporter(app.datastore.FindImagesWithNoAlt, ErrorImagesWithNoAlt)
		rm.addReporter(app.datastore.findRedirectChains, ErrorRedirectChain)
		rm.addReporter(app.datastore.FindPageReportsWithoutH1, ErrorNoH1)
		rm.addReporter(app.datastore.FindPageReportsWithNoLangAttr, ErrorNoLang)
		rm.addReporter(app.datastore.FindPageReportsWithHTTPLinks, ErrorHTTPLinks)
		rm.addReporter(app.datastore.FindMissingHrelangReturnLinks, ErrorHreflangsReturnLink)

		issues := rm.createIssues(cid)
		app.datastore.saveIssues(issues, cid)

		totalIssues := len(issues)

		app.datastore.saveEndIssues(cid, time.Now(), totalIssues)

		fmt.Println("Done.")
	}()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *App) serveIssues(user *User, w http.ResponseWriter, r *http.Request) {
	cid, err := strconv.Atoi(r.URL.Query()["cid"][0])
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

	parsedURL, err := url.Parse(project.URL)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	project.Host = parsedURL.Host

	ig := IssuesGroupView{
		IssuesGroups:    issueGroups,
		Crawl:           crawl,
		Project:         project,
		MediaCount:      mediaCount,
		MediaChart:      mediaChart,
		StatusChart:     statusChart,
		StatusCodeCount: statusCount,
	}

	v := &PageView{
		Data:      ig,
		User:      *user,
		PageTitle: "ISSUES_VIEW",
	}

	renderTemplate(w, "issues", v)
}

func (app *App) serveIssuesView(user *User, w http.ResponseWriter, r *http.Request) {
	eid := r.URL.Query()["eid"][0]
	cid, err := strconv.Atoi(r.URL.Query()["cid"][0])
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	totalPages := app.datastore.getNumberOfPagesForIssues(cid, eid)

	p := r.URL.Query()["p"]
	page := 1
	if len(p) > 0 {
		page, err = strconv.Atoi(r.URL.Query()["p"][0])
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

	parsedURL, err := url.Parse(project.URL)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	project.Host = parsedURL.Host

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

	renderTemplate(w, "issues_view", v)
}

func (app *App) serveResourcesView(user *User, w http.ResponseWriter, r *http.Request) {
	rid, err := strconv.Atoi(r.URL.Query()["rid"][0])
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	cid, err := strconv.Atoi(r.URL.Query()["cid"][0])
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	eid := r.URL.Query()["eid"][0]

	u, err := app.datastore.findCrawlUserId(cid)
	if err != nil || u.Id != user.Id {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	pageReport := app.datastore.FindPageReportById(rid)
	errorTypes := app.datastore.findErrorTypesByPage(rid, cid)
	inLinks := app.datastore.FindInLinks(pageReport.URL, cid)
	redirects := app.datastore.FindPageReportsRedirectingToURL(pageReport.URL, cid)

	rv := ResourcesView{
		PageReport: pageReport,
		Cid:        cid, Eid: eid,
		ErrorTypes: errorTypes,
		InLinks:    inLinks,
		Redirects:  redirects,
	}

	v := &PageView{
		Data:      rv,
		User:      *user,
		PageTitle: "RESOURCES_VIEW",
	}

	renderTemplate(w, "resources", v)
}

func (app *App) serveSignup(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Println(err)
			http.Redirect(w, r, "/signup", http.StatusSeeOther)

			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Println(err)
			http.Redirect(w, r, "/signup", http.StatusSeeOther)

			return
		}

		app.datastore.userSignup(email, string(hashedPassword))

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	v := &PageView{
		PageTitle: "SIGNUP_VIEW",
	}

	renderTemplate(w, "signup", v)
}

func (app *App) serveSignin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Println(err)
			http.Redirect(w, r, "/signin", http.StatusSeeOther)

			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")

		u := app.datastore.findUserByEmail(email)
		if u.Id == 0 {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		if err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		session, _ := app.cookie.Get(r, "SESSION_ID")
		session.Values["authenticated"] = true
		session.Values["uid"] = u.Id
		session.Save(r, w)

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	v := &PageView{
		PageTitle: "SIGNIN_VIEW",
	}

	renderTemplate(w, "signin", v)
}

func (app *App) serveDownloadAll(user *User, w http.ResponseWriter, r *http.Request) {
	cid, err := strconv.Atoi(r.URL.Query()["cid"][0])
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

	crawl := app.datastore.findCrawlById(cid)

	project, err := app.datastore.findProjectById(crawl.ProjectId, user.Id)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	parsedURL, err := url.Parse(project.URL)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	var pageReports []PageReport

	eid := r.URL.Query()["eid"]
	fileName := parsedURL.Host

	if len(eid) > 0 && eid[0] != "" {
		fileName = fileName + "-" + eid[0]
		pageReports = app.datastore.FindAllPageReportsByCrawlIdAndErrorType(cid, eid[0])
	} else {
		pageReports = app.datastore.FindAllPageReportsByCrawlId(cid)
	}

	w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.csv\"", fileName))

	initCSV(w)
	for _, p := range pageReports {
		writeCSVPageReport(p)
	}
}

func (app *App) serveSignout(user *User, w http.ResponseWriter, r *http.Request) {
	session, _ := app.cookie.Get(r, "SESSION_ID")
	session.Values["authenticated"] = false
	session.Values["uid"] = nil
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *App) requireAuth(f func(user *User, w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := app.cookie.Get(r, "SESSION_ID")
		var authenticated interface{} = session.Values["authenticated"]
		if authenticated != nil {
			isAuthenticated := session.Values["authenticated"].(bool)
			if isAuthenticated {
				session, _ := app.cookie.Get(r, "SESSION_ID")
				uid := session.Values["uid"].(int)
				user := app.datastore.findUserById(uid)
				f(user, w, r)

				return
			}
		}

		http.Redirect(w, r, "/signin", http.StatusSeeOther)
	}
}
