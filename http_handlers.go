package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id       int
	Email    string
	Password string
}

type ProjectView struct {
	Project         Project
	Crawl           Crawl
	TotalCount      int
	MediaCount      map[string]int
	StatusCodeCount map[int]int
	TotalIssues     int
}

type IssuesGroupView struct {
	IssuesGroups    map[string]IssueGroup
	Project         Project
	Crawl           Crawl
	TotalCount      int
	MediaCount      CountList
	StatusCodeCount CountList
	TotalIssues     int
	MediaChart      Chart
	StatusChart     Chart
}

type IssuesView struct {
	PageReports []PageReport
	Cid         int
	Eid         string
	Project     Project
}

type ResourcesView struct {
	PageReport PageReport
	Cid        int
	Eid        string
	ErrorTypes []string
	InLinks    []PageReport
	Redirects  []PageReport
}

type Crawl struct {
	Id        int
	ProjectId int
	URL       string
	Start     time.Time
	End       sql.NullTime
}

type Project struct {
	Id              int
	URL             string
	Host            string
	IgnoreRobotsTxt bool
	Created         time.Time
}

var cookie *sessions.CookieStore

func init() {
	cookie = sessions.NewCookieStore([]byte("SESSION_ID"))
}

func (c Crawl) TotalTime() time.Duration {
	if c.End.Valid {
		return c.End.Time.Sub(c.Start)
	}

	return 0
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	session, _ := cookie.Get(r, "SESSION_ID")
	uid := session.Values["uid"].(int)

	var views []ProjectView
	projects := findProjectsByUser(uid)

	for _, p := range projects {
		c := getLastCrawl(&p)
		pv := ProjectView{
			Project:    p,
			Crawl:      c,
			TotalCount: CountCrawled(c.Id),
			//			MediaCount:      CountByMediaType(c.Id),
			//			StatusCodeCount: CountByStatusCode(c.Id),
			TotalIssues: countIssuesByCrawl(c.Id),
		}
		views = append(views, pv)
	}

	u := findUserById(uid)
	v := &PageView{
		Data:      views,
		User:      *u,
		PageTitle: "PROJECTS_VIEW",
	}

	renderTemplate(w, "home", v)
}

func serveProjectAdd(w http.ResponseWriter, r *http.Request) {
	var url string

	session, _ := cookie.Get(r, "SESSION_ID")
	uid := session.Values["uid"].(int)

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
		saveProject(url, ignoreRobotsTxt, uid)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	u := findUserById(uid)
	v := &PageView{
		User:      *u,
		PageTitle: "ADD_PROJECT",
	}

	renderTemplate(w, "project_add", v)
}

func serveCrawl(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query()["pid"][0])
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	session, _ := cookie.Get(r, "SESSION_ID")
	uid := session.Values["uid"].(int)

	p, err := findProjectById(pid, uid)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	fmt.Printf("Crawling %s...\n", p.URL)
	go func() {
		start := time.Now()
		cid := startCrawler(p)
		fmt.Println(time.Since(start))
		fmt.Printf("Creating issues for crawl id %d.\n", cid)
		rm := NewReportManager()
		rm.createIssues(cid)
		fmt.Println("Done.")
	}()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func serveIssues(w http.ResponseWriter, r *http.Request) {
	cid, err := strconv.Atoi(r.URL.Query()["cid"][0])
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	session, _ := cookie.Get(r, "SESSION_ID")
	uid := session.Values["uid"].(int)
	u, err := findCrawlUserId(cid)
	if err != nil || u.Id != uid {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	issueGroups := findIssues(cid)
	crawl := findCrawlById(cid)
	project, err := findProjectById(crawl.ProjectId, uid)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	mediaCount := CountByMediaType(cid)
	mediaChart := NewChart(mediaCount)
	statusCount := CountByStatusCode(cid)
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
		TotalCount:      CountCrawled(cid),
		MediaCount:      mediaCount,
		MediaChart:      mediaChart,
		StatusChart:     statusChart,
		StatusCodeCount: statusCount,
		TotalIssues:     countIssuesByCrawl(cid),
	}

	v := &PageView{
		Data:      ig,
		User:      *u,
		PageTitle: "ISSUES_VIEW",
	}

	renderTemplate(w, "issues", v)
}

func serveIssuesView(w http.ResponseWriter, r *http.Request) {
	eid := r.URL.Query()["eid"][0]
	cid, err := strconv.Atoi(r.URL.Query()["cid"][0])
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	session, _ := cookie.Get(r, "SESSION_ID")
	uid := session.Values["uid"].(int)
	u, err := findCrawlUserId(cid)
	if err != nil || u.Id != uid {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	crawl := findCrawlById(cid)
	project, err := findProjectById(crawl.ProjectId, uid)
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

	issues := findPageReportIssues(cid, eid)

	view := IssuesView{
		Cid:         cid,
		Eid:         eid,
		PageReports: issues,
		Project:     project,
	}

	v := &PageView{
		Data:      view,
		User:      *u,
		PageTitle: "ISSUES_DETAIL",
	}

	renderTemplate(w, "issues_view", v)
}

func serveResourcesView(w http.ResponseWriter, r *http.Request) {
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

	session, _ := cookie.Get(r, "SESSION_ID")
	uid := session.Values["uid"].(int)
	u, err := findCrawlUserId(cid)
	if err != nil || u.Id != uid {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	pageReport := FindPageReportById(rid)
	errorTypes := findErrorTypesByPage(rid, cid)
	inLinks := FindInLinks(pageReport.URL, cid)
	redirects := FindPageReportsRedirectingToURL(pageReport.URL, cid)

	rv := ResourcesView{
		PageReport: pageReport,
		Cid:        cid, Eid: eid,
		ErrorTypes: errorTypes,
		InLinks:    inLinks,
		Redirects:  redirects,
	}

	v := &PageView{
		Data:      rv,
		User:      *u,
		PageTitle: "RESOURCES_VIEW",
	}

	renderTemplate(w, "resources", v)
}

func serveSignup(w http.ResponseWriter, r *http.Request) {
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

		userSignup(email, string(hashedPassword))

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	v := &PageView{
		PageTitle: "SIGNUP_VIEW",
	}

	renderTemplate(w, "signup", v)
}

func serveSignin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Println(err)
			http.Redirect(w, r, "/signin", http.StatusSeeOther)

			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")

		u := findUserByEmail(email)
		if u.Id == 0 {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		if err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		session, _ := cookie.Get(r, "SESSION_ID")
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

func serveSignout(w http.ResponseWriter, r *http.Request) {
	session, _ := cookie.Get(r, "SESSION_ID")
	session.Values["authenticated"] = false
	session.Values["uid"] = nil
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func requireAuth(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := cookie.Get(r, "SESSION_ID")
		var authenticated interface{} = session.Values["authenticated"]
		if authenticated != nil {
			isAuthenticated := session.Values["authenticated"].(bool)
			if isAuthenticated {
				f(w, r)

				return
			}
		}

		http.Redirect(w, r, "/signin", http.StatusSeeOther)
	}
}
