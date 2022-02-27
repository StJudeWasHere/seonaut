package app

import (
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"database/sql"

	"github.com/mnlg/lenkrr/internal/config"
	"github.com/mnlg/lenkrr/internal/issue"
	"github.com/mnlg/lenkrr/internal/project"
	"github.com/mnlg/lenkrr/internal/report"
	"github.com/mnlg/lenkrr/internal/user"

	_ "github.com/go-sql-driver/mysql"
)

type datastore struct {
	db *sql.DB
}

const (
	paginationMax        = 25
	maxOpenConns         = 25
	maxIddleConns        = 25
	connMaxLifeInMinutes = 5
)

func NewDataStore(config config.DBConfig) (*datastore, error) {
	db, err := sql.Open("mysql", fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?parseTime=true",
		config.User,
		config.Pass,
		config.Server,
		config.Port,
		config.Name,
	))

	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIddleConns)
	db.SetConnMaxLifetime(connMaxLifeInMinutes * time.Minute)

	if err := db.Ping(); err != nil {
		log.Printf("Unable to reach database: %v\n", err)
		return nil, err
	}

	return &datastore{db: db}, nil
}

func (ds *datastore) CountCrawled(cid int) int {
	row := ds.db.QueryRow("SELECT count(*) FROM pagereports WHERE crawl_id = ?", cid)
	var c int
	if err := row.Scan(&c); err != nil {
		log.Printf("CountCrawled: %v\n", err)
	}

	return c
}

func (ds *datastore) CountByMediaType(cid int) issue.CountList {
	query := `
		SELECT media_type, count(*)
		FROM pagereports
		WHERE crawl_id = ?
		GROUP BY media_type`

	return ds.countListQuery(query, cid)
}

func (ds *datastore) CountByStatusCode(cid int) issue.CountList {
	query := `
		SELECT
			status_code,
			count(*)
		FROM pagereports
		WHERE crawl_id = ?
		GROUP BY status_code`

	return ds.countListQuery(query, cid)
}

func (ds *datastore) countListQuery(query string, cid int) issue.CountList {
	m := issue.CountList{}
	rows, err := ds.db.Query(query, cid)
	if err != nil {
		log.Println(err)
		return m
	}

	for rows.Next() {
		c := issue.CountItem{}
		err := rows.Scan(&c.Key, &c.Value)
		if err != nil {
			log.Println(err)
			continue
		}
		m = append(m, c)
	}

	sort.Sort(sort.Reverse(m))

	return m
}

func (ds *datastore) EmailExists(email string) bool {
	query := `select exists (select id from users where email = ?)`
	var exists bool
	err := ds.db.QueryRow(query, email).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error checking if email exists '%s' %v", email, err)
	}

	return exists
}

func (ds *datastore) UserSignup(user, password string) {
	query := `INSERT INTO users (email, password) VALUES (?, ?)`
	stmt, _ := ds.db.Prepare(query)
	defer stmt.Close()

	_, err := stmt.Exec(user, password)
	if err != nil {
		log.Printf("WserSignup: %v\n", err)
	}
}

func (ds *datastore) FindUserByEmail(email string) *user.User {
	u := user.User{}
	query := `
		SELECT
			id,
			email,
			password,
			IF (period_end > NOW() is NULL, FALSE, period_end > NOW()) AS advanced,
			stripe_session_id
		FROM users
		WHERE email = ?`

	row := ds.db.QueryRow(query, email)
	err := row.Scan(&u.Id, &u.Email, &u.Password, &u.Advanced, &u.StripeSessionId)
	if err != nil {
		log.Println(err)
		return &u
	}

	return &u
}

func (ds *datastore) FindUserById(id int) *user.User {
	u := user.User{}
	query := `
		SELECT
			id,
			email,
			password,
			IF (period_end > NOW() is NULL, FALSE, period_end > NOW()) AS advanced,
			stripe_session_id
		FROM users
		WHERE id = ?`

	row := ds.db.QueryRow(query, id)
	err := row.Scan(&u.Id, &u.Email, &u.Password, &u.Advanced, &u.StripeSessionId)
	if err != nil {
		log.Println(err)
		return &u
	}

	return &u
}

func (ds *datastore) UserSetStripeId(email, stripeCustomerId string) {
	query := `
		UPDATE users
		SET stripe_customer_id = ?
		WHERE email = ?`

	stmt, _ := ds.db.Prepare(query)
	defer stmt.Close()

	_, err := stmt.Exec(stripeCustomerId, email)
	if err != nil {
		log.Printf("userUpgrade: %v\n", err)
	}
}

func (ds *datastore) UserSetStripeSession(id int, stripeSessionId string) {
	query := `
		UPDATE users
		SET stripe_session_id = ?
		WHERE id = ?`

	stmt, _ := ds.db.Prepare(query)
	defer stmt.Close()

	_, err := stmt.Exec(stripeSessionId, id)
	if err != nil {
		log.Printf("userUpgrade: %v\n", err)
	}
}

func (ds *datastore) RenewSubscription(stripeCustomerId string) {
	query := `
		UPDATE users
		SET period_end = ?
		WHERE stripe_customer_id = ?`

	stmt, _ := ds.db.Prepare(query)
	defer stmt.Close()

	periodEnd := time.Now().AddDate(0, 1, 2)

	_, err := stmt.Exec(periodEnd, stripeCustomerId)
	if err != nil {
		log.Printf("userUpgrade: %v\n", err)
	}
}

func (ds *datastore) findCrawlUserId(cid int) (*user.User, error) {
	u := user.User{}
	query := `
		SELECT 
			users.id,
			users.email,
			users.password
		FROM crawls
		LEFT JOIN projects ON projects.id = crawls.project_id
		LEFT JOIN users ON projects.user_id = users.id
		WHERE crawls.id = ?`

	row := ds.db.QueryRow(query, cid)
	err := row.Scan(&u.Id, &u.Email, &u.Password)
	if err != nil {
		log.Println(err)
		return &u, err
	}

	return &u, nil
}

func (ds *datastore) SaveCrawl(p project.Project) int64 {
	stmt, _ := ds.db.Prepare("INSERT INTO crawls (project_id) VALUES (?)")
	defer stmt.Close()
	res, err := stmt.Exec(p.Id)

	if err != nil {
		log.Printf("saveCrawl\nProject: %+v\nError: %+v\n", p, err)
		return 0
	}

	cid, err := res.LastInsertId()
	if err != nil {
		log.Println(err)
		return 0
	}

	return cid
}

func (ds *datastore) SaveEndCrawl(cid int64, t time.Time, totalURLs int) {
	stmt, _ := ds.db.Prepare("UPDATE crawls SET end = ?, total_urls= ? WHERE id = ?")
	defer stmt.Close()
	_, err := stmt.Exec(t, totalURLs, cid)
	if err != nil {
		log.Printf("saveEndCrawl: %v\n", err)
	}
}

func (ds *datastore) saveEndIssues(cid int, t time.Time, totalIssues int) {
	stmt, _ := ds.db.Prepare("UPDATE crawls SET issues_end = ?, total_issues = ? WHERE id = ?")
	defer stmt.Close()
	_, err := stmt.Exec(t, totalIssues, cid)
	if err != nil {
		log.Printf("saveEndIssues: %v\n", err)
	}
}

func (ds *datastore) getLastCrawl(p *project.Project) Crawl {
	query := `
		SELECT
			id,
			start,
			end,
			total_urls,
			total_issues,
			issues_end
		FROM crawls
		WHERE project_id = ?
		ORDER BY start DESC LIMIT 1`

	row := ds.db.QueryRow(query, p.Id)

	crawl := Crawl{}
	err := row.Scan(&crawl.Id, &crawl.Start, &crawl.End, &crawl.TotalURLs, &crawl.TotalIssues, &crawl.IssuesEnd)
	if err != nil {
		log.Printf("getLastCrawl: %v\n", err)
	}

	return crawl
}

func (ds *datastore) SaveProject(s string, ignoreRobotsTxt, useJavascript bool, uid int) {
	query := `
		INSERT INTO projects (url, ignore_robotstxt, use_javascript, user_id)
		VALUES (?, ?, ?, ?)
	`

	stmt, _ := ds.db.Prepare(query)
	defer stmt.Close()
	_, err := stmt.Exec(s, ignoreRobotsTxt, useJavascript, uid)
	if err != nil {
		log.Printf("saveProject: %v\n", err)
	}
}

func (ds *datastore) FindProjectsByUser(uid int) []project.Project {
	var projects []project.Project
	query := `
		SELECT id, url, ignore_robotstxt, use_javascript, created
		FROM projects
		WHERE user_id = ?`

	rows, err := ds.db.Query(query, uid)
	if err != nil {
		log.Println(err)
		return projects
	}

	for rows.Next() {
		p := project.Project{}
		err := rows.Scan(&p.Id, &p.URL, &p.IgnoreRobotsTxt, &p.UseJS, &p.Created)
		if err != nil {
			log.Println(err)
			continue
		}

		projects = append(projects, p)
	}

	return projects
}

func (ds *datastore) findCrawlById(cid int) Crawl {
	row := ds.db.QueryRow("SELECT id, project_id, start, end FROM crawls WHERE id = ?", cid)

	c := Crawl{}
	err := row.Scan(&c.Id, &c.ProjectId, &c.Start, &c.End)
	if err != nil {
		log.Println(err)
		return c
	}

	return c
}

func (ds *datastore) FindProjectById(id int, uid int) (project.Project, error) {
	query := `
		SELECT id, url, ignore_robotstxt, use_javascript, created
		FROM projects
		WHERE id = ? AND user_id = ?`

	row := ds.db.QueryRow(query, id, uid)

	p := project.Project{}
	err := row.Scan(&p.Id, &p.URL, &p.IgnoreRobotsTxt, &p.UseJS, &p.Created)
	if err != nil {
		log.Println(err)
		return p, err
	}

	return p, nil
}

func (ds *datastore) saveIssues(issues []Issue, cid int) {
	query := `
		INSERT INTO issues (pagereport_id, crawl_id, issue_type_id)
		VALUES (?, ?, ?)`

	stmt, _ := ds.db.Prepare(query)
	defer stmt.Close()

	for _, i := range issues {
		_, err := stmt.Exec(i.PageReportId, cid, i.ErrorType)
		if err != nil {
			log.Printf("saveIssues -> ID: %d ERROR: %d CRAWL: %d %v\n", i.PageReportId, i.ErrorType, cid, err)
			continue
		}
	}
}

func (ds *datastore) FindIssues(cid int) map[string]issue.IssueGroup {
	issues := map[string]issue.IssueGroup{}
	query := `
		SELECT
			issue_types.type,
			issue_types.priority,
			count(DISTINCT issues.pagereport_id)
		FROM issues
		INNER JOIN  issue_types ON issue_types.id = issues.issue_type_id
		WHERE crawl_id = ? GROUP BY issue_type_id`

	rows, err := ds.db.Query(query, cid)
	if err != nil {
		log.Println(err)
		return issues
	}

	for rows.Next() {
		ig := issue.IssueGroup{}
		err := rows.Scan(&ig.ErrorType, &ig.Priority, &ig.Count)
		if err != nil {
			log.Println(err)
			continue
		}

		issues[ig.ErrorType] = ig
	}

	return issues
}

func (ds *datastore) countIssuesByCrawl(cid int) int {
	var c int
	row := ds.db.QueryRow("SELECT count(*) FROM issues WHERE crawl_id = ?", cid)
	if err := row.Scan(&c); err != nil {
		log.Println(err)
	}

	return c
}

func (ds *datastore) findErrorTypesByPage(pid, cid int) []string {
	var et []string
	query := `
		SELECT 
			issue_types.type
		FROM issues
		INNER JOIN issue_types ON issue_types.id = issues.issue_type_id
		WHERE pagereport_id = ? and crawl_id = ?
		GROUP BY issue_type_id`

	rows, err := ds.db.Query(query, pid, cid)
	if err != nil {
		log.Println(err)
		return et
	}

	for rows.Next() {
		var s string
		err := rows.Scan(&s)
		if err != nil {
			log.Println(err)
			continue
		}
		et = append(et, s)
	}

	return et
}

func (ds *datastore) SavePageReport(r *report.PageReport, cid int64) {
	urlHash := hash(r.URL)
	var redirectHash string
	if r.RedirectURL != "" {
		redirectHash = hash(r.RedirectURL)
	}

	query := `
		INSERT INTO pagereports (
			crawl_id,
			url,
			url_hash,
			scheme,
			redirect_url,
			redirect_hash,
			refresh,
			status_code,
			content_type,
			media_type,
			lang,
			title,
			description,
			robots,
			canonical,
			h1,
			h2,
			words,
			size,
			valid_headings
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	stmt, err := ds.db.Prepare(query)
	if err != nil {
		log.Printf("saveReport: %v\n", err)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(
		cid,
		r.URL,
		urlHash,
		r.ParsedURL.Scheme,
		r.RedirectURL,
		redirectHash,
		r.Refresh,
		r.StatusCode,
		r.ContentType,
		r.MediaType,
		r.Lang,
		r.Title,
		r.Description,
		r.Robots,
		r.Canonical,
		r.H1,
		r.H2,
		r.Words,
		len(r.Body),
		r.ValidHeadings,
	)

	if err != nil {
		log.Printf("Error in SavePageReport\nCID: %v\n Report: %+v\nError: %+v\n", cid, r, err)
		return
	}

	lid, err := res.LastInsertId()
	if err != nil {
		log.Println(err)
		return
	}

	if len(r.Links) > 0 {
		sqlString := "INSERT INTO links (pagereport_id, crawl_id, url, scheme, rel, nofollow, text, url_hash) values "
		v := []interface{}{}
		for _, l := range r.Links {
			hash := hash(l.URL)
			sqlString += "(?, ?, ?, ?, ?, ?, ?, ?),"
			v = append(v, lid, cid, l.URL, l.ParsedURL.Scheme, l.Rel, l.NoFollow, l.Text, hash)
		}
		sqlString = sqlString[0 : len(sqlString)-1]
		stmt, err := ds.db.Prepare(sqlString)
		if err != nil {
			log.Printf("saveReport links: %v\n", err)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(v...)
		if err != nil {
			log.Printf("Error in SavePageReport\nCID: %v\n Links: %+v\nError: %+v\n", cid, v, err)
		}
	}

	if len(r.ExternalLinks) > 0 {
		sqlString := "INSERT INTO external_links (pagereport_id, crawl_id, url, rel, nofollow, text) values "
		v := []interface{}{}
		for _, l := range r.ExternalLinks {
			sqlString += "(?, ?, ?, ?, ?, ?),"
			v = append(v, lid, cid, l.URL, l.Rel, l.NoFollow, l.Text)
		}
		sqlString = sqlString[0 : len(sqlString)-1]
		stmt, err := ds.db.Prepare(sqlString)
		if err != nil {
			log.Println(err)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(v...)
		if err != nil {
			log.Printf("Error in SavePageReport\nCID: %v\n Links: %+v\nError: %+v\n", cid, v, err)
		}
	}

	if len(r.Hreflangs) > 0 {
		sqlString := "INSERT INTO hreflangs (pagereport_id, crawl_id, from_lang, to_url, to_lang, from_hash, to_hash) values "
		v := []interface{}{}
		for _, h := range r.Hreflangs {
			sqlString += "(?, ?, ?, ?, ?, ?, ?),"
			v = append(v, lid, cid, r.Lang, h.URL, h.Lang, hash(r.URL), hash(h.URL))
		}
		sqlString = sqlString[0 : len(sqlString)-1]
		stmt, _ := ds.db.Prepare(sqlString)
		defer stmt.Close()

		_, err := stmt.Exec(v...)
		if err != nil {
			log.Printf("savePageReport\nCID: %v\n Hreflangs: %+v\nError: %+v\n", cid, v, err)
		}
	}

	if len(r.Images) > 0 {
		sqlString := "INSERT INTO images (pagereport_id, url, alt, crawl_id) values "
		v := []interface{}{}
		for _, i := range r.Images {
			sqlString += "(?, ?, ?, ?),"
			v = append(v, lid, i.URL, i.Alt, cid)
		}
		sqlString = sqlString[0 : len(sqlString)-1]
		stmt, _ = ds.db.Prepare(sqlString)
		defer stmt.Close()

		_, err := stmt.Exec(v...)
		if err != nil {
			log.Printf("savePageReport\nCID: %v\n Images: %+v\nError: %+v\n", cid, v, err)
		}
	}

	if len(r.Scripts) > 0 {
		sqlString := "INSERT INTO scripts (pagereport_id, url, crawl_id) values "
		v := []interface{}{}
		for _, s := range r.Scripts {
			sqlString += "(?, ?, ?),"
			v = append(v, lid, s, cid)
		}
		sqlString = sqlString[0 : len(sqlString)-1]
		stmt, _ := ds.db.Prepare(sqlString)
		defer stmt.Close()

		_, err := stmt.Exec(v...)
		if err != nil {
			log.Printf("savePageReport\nCID: %v\n Scripts: %+v\nError: %+v\n", cid, v, err)
		}
	}

	if len(r.Styles) > 0 {
		sqlString := "INSERT INTO styles (pagereport_id, url, crawl_id) values "
		v := []interface{}{}

		for _, s := range r.Styles {
			sqlString += "(?, ?, ?),"
			v = append(v, lid, s, cid)

		}
		sqlString = sqlString[0 : len(sqlString)-1]
		stmt, _ := ds.db.Prepare(sqlString)
		defer stmt.Close()

		_, err := stmt.Exec(v...)
		if err != nil {
			log.Printf("savePageReport\nCID: %v\n Styles: %+v\nError: %+v\n", cid, v, err)
		}
	}
}

func (ds *datastore) FindAllPageReportsByCrawlId(cid int) []report.PageReport {
	var pageReports []report.PageReport
	query := `
		SELECT
			id,
			url,
			redirect_url,
			refresh,
			status_code,
			content_type,
			media_type,
			lang,
			title,
			description,
			robots,
			canonical,
			h1,
			h2,
			words,
			size,
			valid_headings
		FROM pagereports
		WHERE crawl_id = ?`

	rows, err := ds.db.Query(query, cid)
	if err != nil {
		log.Println(err)
	}

	for rows.Next() {
		p := report.PageReport{}
		err := rows.Scan(&p.Id,
			&p.URL,
			&p.RedirectURL,
			&p.Refresh,
			&p.StatusCode,
			&p.ContentType,
			&p.MediaType,
			&p.Lang,
			&p.Title,
			&p.Description,
			&p.Robots,
			&p.Canonical,
			&p.H1,
			&p.H2,
			&p.Words,
			&p.Size,
			&p.ValidHeadings,
		)
		if err != nil {
			log.Println(err)
			continue
		}

		pageReports = append(pageReports, p)
	}

	return pageReports
}

func (ds *datastore) FindAllPageReportsByCrawlIdAndErrorType(cid int, et string) []report.PageReport {
	var pageReports []report.PageReport
	query := `
		SELECT
			id,
			url,
			redirect_url,
			refresh,
			status_code,
			content_type,
			media_type,
			lang,
			title,
			description,
			robots,
			canonical,
			h1,
			h2,
			words,
			size,
			valid_headings
		FROM pagereports
		WHERE crawl_id = ?
		AND id IN (
			SELECT
				pagereport_id 
			FROM issues
			INNER JOIN issue_types ON issue_types.id = issues.issue_type_id
			WHERE issue_types.type = ? AND crawl_id = ?
		)`

	rows, err := ds.db.Query(query, cid, et, cid)
	if err != nil {
		log.Println(err)
	}

	for rows.Next() {
		p := report.PageReport{}
		err := rows.Scan(&p.Id,
			&p.URL,
			&p.RedirectURL,
			&p.Refresh,
			&p.StatusCode,
			&p.ContentType,
			&p.MediaType,
			&p.Lang,
			&p.Title,
			&p.Description,
			&p.Robots,
			&p.Canonical,
			&p.H1,
			&p.H2,
			&p.Words,
			&p.Size,
			&p.ValidHeadings,
		)
		if err != nil {
			log.Println(err)
			continue
		}

		pageReports = append(pageReports, p)
	}

	return pageReports
}

func (ds *datastore) FindPageReportById(rid int) report.PageReport {
	query := `
		SELECT
			id,
			url,
			redirect_url,
			refresh,
			status_code,
			content_type,
			media_type,
			lang,
			title,
			description,
			robots,
			canonical,
			h1,
			h2,
			words,
			size,
			valid_headings
		FROM pagereports
		WHERE id = ?`

	row := ds.db.QueryRow(query, rid)

	p := report.PageReport{}
	err := row.Scan(&p.Id,
		&p.URL,
		&p.RedirectURL,
		&p.Refresh,
		&p.StatusCode,
		&p.ContentType,
		&p.MediaType,
		&p.Lang,
		&p.Title,
		&p.Description,
		&p.Robots,
		&p.Canonical,
		&p.H1,
		&p.H2,
		&p.Words,
		&p.Size,
		&p.ValidHeadings,
	)
	if err != nil {
		log.Println(err)
	}

	lrows, err := ds.db.Query("SELECT url, rel, nofollow, text FROM links WHERE pagereport_id = ? limit 25", rid)
	if err != nil {
		log.Println(err)
	}

	for lrows.Next() {
		l := report.Link{}
		err = lrows.Scan(&l.URL, &l.Rel, &l.NoFollow, &l.Text)
		if err != nil {
			log.Println(err)
			continue
		}

		p.Links = append(p.Links, l)
	}

	lrows, err = ds.db.Query("SELECT url, rel, nofollow, text FROM external_links WHERE pagereport_id = ? limit 25", rid)
	if err != nil {
		log.Println(err)
	}

	for lrows.Next() {
		l := report.Link{}
		err = lrows.Scan(&l.URL, &l.Rel, &l.NoFollow, &l.Text)
		if err != nil {
			log.Println(err)
			continue
		}

		p.ExternalLinks = append(p.ExternalLinks, l)
	}

	hrows, err := ds.db.Query("SELECT to_url, to_lang FROM hreflangs WHERE pagereport_id = ?", rid)
	if err != nil {
		log.Println(err)
	}

	for hrows.Next() {
		h := report.Hreflang{}
		err = hrows.Scan(&h.URL, &h.Lang)
		if err != nil {
			log.Println(err)
			continue
		}

		p.Hreflangs = append(p.Hreflangs, h)
	}

	irows, err := ds.db.Query("SELECT url, alt FROM images WHERE pagereport_id = ?", rid)
	if err != nil {
		log.Println(err)
	}

	for irows.Next() {
		i := report.Image{}
		err = irows.Scan(&i.URL, &i.Alt)
		if err != nil {
			log.Println(err)
			continue
		}

		p.Images = append(p.Images, i)
	}

	scrows, err := ds.db.Query("SELECT url FROM scripts WHERE pagereport_id = ?", rid)
	if err != nil {
		log.Println(err)
	}

	for scrows.Next() {
		var url string
		err = scrows.Scan(&url)
		if err != nil {
			log.Println(err)
			continue
		}

		p.Scripts = append(p.Scripts, url)
	}

	strows, err := ds.db.Query("SELECT url FROM styles WHERE pagereport_id = ?", rid)
	if err != nil {
		log.Println(err)
	}

	for strows.Next() {
		var url string
		err = strows.Scan(&url)
		if err != nil {
			log.Println(err)
			continue
		}

		p.Styles = append(p.Styles, url)
	}

	return p
}

func (ds *datastore) FindPreviousCrawlId(pid int) int {
	query := `
		SELECT
			id
		FROM crawls
		WHERE project_id = ?
		ORDER BY end DESC
		LIMIT 1, 1`

	row := ds.db.QueryRow(query, pid)
	var c int
	if err := row.Scan(&c); err != nil {
		log.Printf("FindPreviousCrawlId: %v\n", err)
	}

	return c
}

func (ds *datastore) DeletePreviousCrawl(pid int) {
	previousCrawl := ds.FindPreviousCrawlId(pid)

	var deleteFunc func(cid int, table string)
	deleteFunc = func(cid int, table string) {
		query := fmt.Sprintf("DELETE FROM %s WHERE crawl_id = ? ORDER BY id DESC LIMIT 1000", table)
		_, err := ds.db.Exec(query, previousCrawl)
		if err != nil {
			log.Printf("DeletePreviousCeawl: pid %d table %s %v\n", pid, table, err)
			return
		}

		query = fmt.Sprintf("SELECT count(*) FROM %s WHERE crawl_id = ?", table)
		row := ds.db.QueryRow(query, previousCrawl)
		var c int
		if err := row.Scan(&c); err != nil {
			log.Printf("DeletePreviousCrawl count: pid %d table %s %v\n", pid, table, err)
		}

		if c > 0 {
			time.Sleep(1500 * time.Millisecond)
			deleteFunc(cid, table)
		}
	}

	deleteFunc(previousCrawl, "links")
	deleteFunc(previousCrawl, "external_links")
	deleteFunc(previousCrawl, "hreflangs")
	deleteFunc(previousCrawl, "issues")
	deleteFunc(previousCrawl, "images")
	deleteFunc(previousCrawl, "scripts")
	deleteFunc(previousCrawl, "styles")
	deleteFunc(previousCrawl, "pagereports")
}

func (ds *datastore) getNumberOfPagesForIssues(cid int, errorType string) int {
	query := `
		SELECT count(DISTINCT pagereport_id)
		FROM issues
		INNER JOIN issue_types ON issue_types.id = issues.issue_type_id
		WHERE issue_types.type = ? and crawl_id  = ?`

	row := ds.db.QueryRow(query, errorType, cid)
	var c int
	if err := row.Scan(&c); err != nil {
		log.Printf("CountCrawled: %v\n", err)
	}
	var f float64 = float64(c) / float64(paginationMax)
	return int(math.Ceil(f))
}
