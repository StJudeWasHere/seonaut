package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

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

type PageReportView struct {
	Projects              []Project
	Crawl                 Crawl
	Error30x              []PageReport
	Error40x              []PageReport
	Error50x              []PageReport
	LittleContent         []PageReport
	ImagesWithNoAlt       []PageReport
	PageReports           []PageReport
	EmptyTitle            []PageReport
	ShortTitle            []PageReport
	LongTitle             []PageReport
	DuplicatedTitle       []PageReport
	EmptyDescription      []PageReport
	ShortDescription      []PageReport
	LongDescription       []PageReport
	DuplicatedDescription []PageReport
	TotalCount            int
	MediaCount            map[string]int
	StatusCodeCount       map[int]int
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

func (c Crawl) TotalTime() time.Duration {
	return c.End.Sub(c.Start)
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

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

	var templates = template.Must(template.ParseFiles(
		"templates/home.html", "templates/head.html", "templates/footer.html", "templates/list.html",
	))

	err := templates.ExecuteTemplate(w, "home.html", views)
	if err != nil {
		fmt.Println(err)
	}
}

func serveProjectAdd(w http.ResponseWriter, r *http.Request) {
	var url string

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			fmt.Println(err)
		}
		url = r.FormValue("url")
		saveProject(url)

	}

	var templates = template.Must(template.ParseFiles(
		"templates/project_add.html", "templates/head.html", "templates/footer.html",
	))

	err := templates.ExecuteTemplate(w, "project_add.html", struct{ URL string }{URL: url})
	if err != nil {
		fmt.Println(err)
	}
}

func serveCrawl(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(r.URL.Query()["pid"][0])
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	p := findProjectById(pid)
	fmt.Printf("Crawling %s...\n", p.URL)
	go func() {
		start := time.Now()
		startCrawler(p)
		fmt.Println(time.Since(start))
	}()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func serveIssues(w http.ResponseWriter, r *http.Request) {
	cid, err := strconv.Atoi(r.URL.Query()["cid"][0])
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	issueGroups := findIssues(cid)

	var templates = template.Must(template.ParseFiles(
		"templates/issues.html", "templates/head.html", "templates/footer.html",
	))

	err = templates.ExecuteTemplate(w, "issues.html", IssuesGroupView{IssuesGroups: issueGroups, Cid: cid})
	if err != nil {
		fmt.Println(err)
	}
}

func serveIssuesView(w http.ResponseWriter, r *http.Request) {
	eid := r.URL.Query()["eid"][0]
	cid, err := strconv.Atoi(r.URL.Query()["cid"][0])
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	var templates = template.Must(template.ParseFiles(
		"templates/issues_view.html", "templates/head.html", "templates/footer.html",
	))

	issues := findPageReportIssues(cid, eid)

	view := IssuesView{
		Cid:         cid,
		PageReports: issues,
	}

	err = templates.ExecuteTemplate(w, "issues_view.html", view)
	if err != nil {
		fmt.Println(err)
	}
}

func serveResourcesView(w http.ResponseWriter, r *http.Request) {
	rid, err := strconv.Atoi(r.URL.Query()["rid"][0])
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	cid, err := strconv.Atoi(r.URL.Query()["cid"][0])
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	pageReport := FindPageReportById(rid)

	var templates = template.Must(template.ParseFiles(
		"templates/resources.html", "templates/head.html", "templates/footer.html", "templates/pagereport.html",
	))

	err = templates.ExecuteTemplate(w, "resources.html", ResourcesView{PageReport: pageReport, Cid: cid})
	if err != nil {
		fmt.Println(err)
	}
}

func serveSignup(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Println(err)
			http.Redirect(w, r, "/signup", http.StatusSeeOther)
		}

		email := r.FormValue("email")
		password := r.FormValue("password")
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Println(err)
			http.Redirect(w, r, "/signup", http.StatusSeeOther)
		}

		userSignup(email, string(hashedPassword))

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	var templates = template.Must(template.ParseFiles(
		"templates/signup.html", "templates/head.html", "templates/footer.html",
	))

	err := templates.ExecuteTemplate(w, "signup.html", struct{}{})
	if err != nil {
		log.Println(err)
	}
}

func serveSignin(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Println(err)
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
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
			// w.WriteHeader(http.StatusUnauthorized)
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var templates = template.Must(template.ParseFiles(
		"templates/signin.html", "templates/head.html", "templates/footer.html",
	))

	err := templates.ExecuteTemplate(w, "signin.html", struct{}{})
	if err != nil {
		log.Println(err)
	}
}
