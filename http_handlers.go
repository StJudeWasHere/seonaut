package main

import (
	"fmt"
	"log"
	"net/http"
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
}

type IssuesGroupView struct {
	IssuesGroups map[string]IssueGroup
	Cid          int
}

type IssuesView struct {
	PageReports []PageReport
	Cid         int
}

type ResourcesView struct {
	PageReport PageReport
	Cid        int
}

type Crawl struct {
	Id    int
	URL   string
	Start time.Time
	End   time.Time
}

type Project struct {
	Id      int
	URL     string
	Created time.Time
}

var cookie *sessions.CookieStore

func init() {
	cookie = sessions.NewCookieStore([]byte("SESSION_ID"))
}

func (c Crawl) TotalTime() time.Duration {
	return c.End.Sub(c.Start)
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var views []ProjectView
	projects := findProjects()

	for _, p := range projects {
		c := getLastCrawl(&p)
		pv := ProjectView{
			Project:         p,
			Crawl:           c,
			TotalCount:      CountCrawled(c.Id),
			MediaCount:      CountByMediaType(c.Id),
			StatusCodeCount: CountByStatusCode(c.Id),
		}
		views = append(views, pv)
	}

	renderTemplate(w, "home", views)
}

func serveProjectAdd(w http.ResponseWriter, r *http.Request) {
	var url string

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}
		url = r.FormValue("url")
		saveProject(url)

	}

	renderTemplate(w, "project_add", struct{ URL string }{URL: url})
}

func serveCrawl(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query()["pid"][0])
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	p := findProjectById(pid)
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

	issueGroups := findIssues(cid)

	renderTemplate(w, "issues", IssuesGroupView{IssuesGroups: issueGroups, Cid: cid})
}

func serveIssuesView(w http.ResponseWriter, r *http.Request) {
	eid := r.URL.Query()["eid"][0]
	cid, err := strconv.Atoi(r.URL.Query()["cid"][0])
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	issues := findPageReportIssues(cid, eid)

	view := IssuesView{
		Cid:         cid,
		PageReports: issues,
	}

	renderTemplate(w, "issues_view", view)
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

	pageReport := FindPageReportById(rid)

	renderTemplate(w, "resources", ResourcesView{PageReport: pageReport, Cid: cid})
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

	renderTemplate(w, "signup", struct{}{})
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

		fmt.Println(email, password)

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
		session.Save(r, w)

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	renderTemplate(w, "sigin", struct{}{})
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
