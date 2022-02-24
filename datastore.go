package main

import (
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"database/sql"

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

type IssueGroup struct {
	ErrorType string
	Count     int
}

func NewDataStore(config DBConfig) (*datastore, error) {
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

func (ds *datastore) CountByMediaType(cid int) CountList {
	query := `
		SELECT media_type, count(*)
		FROM pagereports
		WHERE crawl_id = ?
		GROUP BY media_type`

	return ds.countListQuery(query, cid)
}

func (ds *datastore) CountByStatusCode(cid int) CountList {
	query := `
		SELECT
			status_code,
			count(*)
		FROM pagereports
		WHERE crawl_id = ?
		GROUP BY status_code`

	return ds.countListQuery(query, cid)
}

func (ds *datastore) countListQuery(query string, cid int) CountList {
	m := CountList{}
	rows, err := ds.db.Query(query, cid)
	if err != nil {
		log.Println(err)
		return m
	}

	for rows.Next() {
		c := CountItem{}
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

func (ds *datastore) emailExists(email string) bool {
	query := `select exists (select id from users where email = ?)`
	var exists bool
	err := ds.db.QueryRow(query, email).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error checking if email exists '%s' %v", email, err)
	}

	return exists
}

func (ds *datastore) userSignup(user, password string) {
	query := `INSERT INTO users (email, password) VALUES (?, ?)`
	stmt, _ := ds.db.Prepare(query)
	defer stmt.Close()

	_, err := stmt.Exec(user, password)
	if err != nil {
		log.Printf("userSignup: %v\n", err)
	}
}

func (ds *datastore) findUserByEmail(email string) *User {
	u := User{}
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

func (ds *datastore) findUserById(id int) *User {
	u := User{}
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

func (ds *datastore) userSetStripeId(email, stripeCustomerId string) {
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

func (ds *datastore) userSetStripeSession(id int, stripeSessionId string) {
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

func (ds *datastore) renewSubscription(stripeCustomerId string) {
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

func (ds *datastore) findCrawlUserId(cid int) (*User, error) {
	u := User{}
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

func (ds *datastore) saveCrawl(p Project) int64 {
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

func (ds *datastore) saveEndCrawl(cid int64, t time.Time, totalURLs int) {
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

func (ds *datastore) getLastCrawl(p *Project) Crawl {
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

func (ds *datastore) saveProject(s string, ignoreRobotsTxt, useJavascript bool, uid int) {
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

func (ds *datastore) findProjectsByUser(uid int) []Project {
	var projects []Project
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
		p := Project{}
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

func (ds *datastore) findProjectById(id int, uid int) (Project, error) {
	query := `
		SELECT id, url, ignore_robotstxt, use_javascript, created
		FROM projects
		WHERE id = ? AND user_id = ?`

	row := ds.db.QueryRow(query, id, uid)

	p := Project{}
	err := row.Scan(&p.Id, &p.URL, &p.IgnoreRobotsTxt, &p.UseJS, &p.Created)
	if err != nil {
		log.Println(err)
		return p, err
	}

	return p, nil
}

func (ds *datastore) saveIssues(issues []Issue, cid int) {
	query := `
		INSERT INTO issues (pagereport_id, crawl_id, error_type)
		VALUES (?, ?, ?)`

	stmt, _ := ds.db.Prepare(query)
	defer stmt.Close()

	for _, i := range issues {
		_, err := stmt.Exec(i.PageReportId, cid, i.ErrorType)
		if err != nil {
			log.Printf("saveIssues -> ID: %d ERROR: %s CRAWL: %d %v\n", i.PageReportId, i.ErrorType, cid, err)
			continue
		}
	}
}

func (ds *datastore) findIssues(cid int) map[string]IssueGroup {
	issues := map[string]IssueGroup{}
	query := `
		SELECT
			error_type,
			count(*)
		FROM issues
		WHERE crawl_id = ? GROUP BY error_type`

	rows, err := ds.db.Query(query, cid)
	if err != nil {
		log.Println(err)
		return issues
	}

	for rows.Next() {
		ig := IssueGroup{}
		err := rows.Scan(&ig.ErrorType, &ig.Count)
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
	query := `SELECT error_type FROM issues WHERE pagereport_id = ? and crawl_id = ? GROUP BY error_type`
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

func (ds *datastore) savePageReport(r *PageReport, cid int64) {
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
		r.parsedURL.Scheme,
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
			v = append(v, lid, cid, l.URL, l.parsedUrl.Scheme, l.Rel, l.NoFollow, l.Text, hash)
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

func (ds *datastore) FindAllPageReportsByCrawlId(cid int) []PageReport {
	var pageReports []PageReport
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
		p := PageReport{}
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

func (ds *datastore) FindAllPageReportsByCrawlIdAndErrorType(cid int, et string) []PageReport {
	var pageReports []PageReport
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
		AND id in (SELECT pagereport_id FROM issues WHERE error_type = ? AND crawl_id = ?)`

	rows, err := ds.db.Query(query, cid, et, cid)
	if err != nil {
		log.Println(err)
	}

	for rows.Next() {
		p := PageReport{}
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

func (ds *datastore) FindPageReportById(rid int) PageReport {
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

	p := PageReport{}
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
		l := Link{}
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
		l := Link{}
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
		h := Hreflang{}
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
		i := Image{}
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
		ORDER BY issues_end DESC
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

func (ds *datastore) FindPageReportsRedirectingToURL(u string, cid int) []PageReport {
	uh := hash(u)
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE redirect_hash = ? AND crawl_id = ?`

	return ds.pageReportsQuery(query, uh, cid)
}

func (ds *datastore) FindPageReportsWithEmptyTitle(cid int) []PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE (title = "" OR title IS NULL) AND media_type = "text/html"
		AND status_code >=200 AND status_code < 300 AND crawl_id = ?`

	return ds.pageReportsQuery(query, cid)
}

func (ds *datastore) Find40xPageReports(cid int) []PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE status_code >= 400 AND status_code < 500 AND crawl_id = ?`

	return ds.pageReportsQuery(query, cid)
}

func (ds *datastore) Find30xPageReports(cid int) []PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE status_code >= 300 AND status_code < 400 AND crawl_id = ?`

	return ds.pageReportsQuery(query, cid)
}

func (ds *datastore) Find50xPageReports(cid int) []PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE status_code >= 500 AND crawl_id = ?`

	return ds.pageReportsQuery(query, cid)
}

func (ds *datastore) FindPageReportsWithLittleContent(cid int) []PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE words < 200 AND status_code >= 200 AND status_code < 300 AND media_type = "text/html" AND crawl_id = ?`

	return ds.pageReportsQuery(query, cid)
}

func (ds *datastore) FindPageReportsWithShortTitle(cid int) []PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE length(title) > 0 AND length(title) < 20 AND media_type = "text/html" AND crawl_id = ?`

	return ds.pageReportsQuery(query, cid)
}

func (ds *datastore) FindPageReportsWithLongTitle(cid int) []PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE length(title) > 60 AND media_type = "text/html" AND crawl_id = ?`

	return ds.pageReportsQuery(query, cid)
}

func (ds *datastore) FindPageReportsWithDuplicatedTitle(cid int) []PageReport {
	query := `
		SELECT
			y.id,
			y.url,
			y.title
		FROM pagereports y
		INNER JOIN (
			SELECT
				title,
				lang,
				count(*) AS c
			FROM pagereports
			WHERE crawl_id = ? AND media_type = "text/html" AND status_code >= 200
			AND status_code < 300 AND (canonical = "" OR canonical = url)
			GROUP BY title, lang
			HAVING c > 1
		) d 
		ON d.title = y.title
		WHERE media_type = "text/html" AND length(y.title) > 0 AND crawl_id = ?
		AND status_code >= 200 AND status_code < 300 AND (canonical = "" OR canonical = url)`

	return ds.pageReportsQuery(query, cid, cid)
}

func (ds *datastore) FindPageReportsWithoutH1(cid int) []PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE (h1 = "" OR h1 IS NULL) AND media_type = "text/html"
		AND status_code >= 200 AND status_code < 300 AND crawl_id = ?`

	return ds.pageReportsQuery(query, cid)
}

func (ds *datastore) FindPageReportsWithEmptyDescription(cid int) []PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE (description = "" OR description IS NULL) AND media_type = "text/html"
		AND status_code >= 200 AND status_code < 300 AND crawl_id = ?`

	return ds.pageReportsQuery(query, cid)
}

func (ds *datastore) FindPageReportsWithShortDescription(cid int) []PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE length(description) > 0 AND length(description) < 80 AND media_type = "text/html" AND crawl_id = ?`

	return ds.pageReportsQuery(query, cid)
}

func (ds *datastore) FindPageReportsWithLongDescription(cid int) []PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE length(description) > 160 AND media_type = "text/html" AND crawl_id = ?`

	return ds.pageReportsQuery(query, cid)
}

func (ds *datastore) FindPageReportsWithDuplicatedDescription(cid int) []PageReport {
	query := `
		SELECT
			y.id,
			y.url,
			y.title
		FROM pagereports y
		INNER JOIN (
			SELECT
				description,
				lang,
				count(*) AS c
			FROM pagereports
			WHERE crawl_id = ? AND media_type = "text/html" AND status_code >= 200
			AND status_code < 300 AND (canonical = "" OR canonical = url)
			GROUP BY description, lang
			HAVING c > 1
		) d 
		ON d.description = y.description
		WHERE y.media_type = "text/html" AND length(y.description) > 0 AND y.crawl_id = ?
		AND status_code >= 200 AND status_code < 300 AND (canonical = "" OR canonical = url`

	return ds.pageReportsQuery(query, cid, cid)
}

func (ds *datastore) FindImagesWithNoAlt(cid int) []PageReport {
	query := `
		SELECT
			pagereports.id,
			pagereports.url,
			pagereports.title
		FROM pagereports
		LEFT JOIN images ON images.pagereport_id = pagereports.id
		WHERE images.alt = "" AND pagereports.crawl_id = ?
		GROUP BY pagereports.id`

	return ds.pageReportsQuery(query, cid)
}

func (ds *datastore) FindPageReportsWithNoLangAttr(cid int) []PageReport {
	query := `
		SELECT
			pagereports.id,
			pagereports.url,
			pagereports.title
		FROM pagereports
		WHERE (pagereports.lang = "" OR pagereports.lang = null) and media_type = "text/html"
		AND pagereports.crawl_id = ?`

	return ds.pageReportsQuery(query, cid)
}

func (ds *datastore) FindPageReportsWithHTTPLinks(cid int) []PageReport {
	query := `
		SELECT
			pagereports.id,
			pagereports.url,
			pagereports.title
		FROM pagereports
		LEFT JOIN links ON links.pagereport_id = pagereports.id
		WHERE pagereports.scheme = "https" AND links.scheme = "http"
		AND pagereports.crawl_id = ?
		GROUP BY links.pagereport_id
		HAVING count(links.pagereport_id) > 1`

	return ds.pageReportsQuery(query, cid)
}

func (ds *datastore) FindMissingHrelangReturnLinks(cid int) []PageReport {
	query := `
		SELECT
			distinct pagereports.id,
			pagereports.URL,
			pagereports.Title
		FROM hreflangs
		LEFT JOIN hreflangs b ON hreflangs.crawl_id = b.crawl_id and hreflangs.from_hash = b.to_hash
		LEFT JOIN pagereports ON hreflangs.pagereport_id = pagereports.id
		WHERE  hreflangs.crawl_id = ? AND hreflangs.to_lang != "x-default"
		AND pagereports.status_code < 300 AND b.id IS NULL
		AND (pagereports.canonical = "" OR pagereports.canonical = pagereports.URL)`

	return ds.pageReportsQuery(query, cid)
}

func (ds *datastore) FindInLinks(s string, cid int) []PageReport {
	hash := hash(s)
	query := `
		SELECT 
			pagereports.id,
			pagereports.url,
			pagereports.Title
		FROM links
		LEFT JOIN pagereports ON pagereports.id = links.pagereport_id
		WHERE links.url_hash = ? AND pagereports.crawl_id = ?
		GROUP BY pagereports.id
		LIMIT 25`

	return ds.pageReportsQuery(query, hash, cid)
}

func (ds *datastore) getNumberOfPagesForIssues(cid int, errorType string) int {
	query := `
		SELECT count(*)
		FROM issues
		WHERE error_type = ? and crawl_id  = ?`

	row := ds.db.QueryRow(query, errorType, cid)
	var c int
	if err := row.Scan(&c); err != nil {
		log.Printf("CountCrawled: %v\n", err)
	}
	var f float64 = float64(c) / float64(paginationMax)
	return int(math.Ceil(f))
}

func (ds *datastore) findPageReportIssues(cid, p int, errorType string) []PageReport {
	max := paginationMax
	offset := max * p

	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE id IN (
			SELECT pagereport_id
			FROM issues
			WHERE error_type = ? and crawl_id  = ?
		) LIMIT ?, ?`

	return ds.pageReportsQuery(query, errorType, cid, offset, max)
}

func (ds *datastore) findRedirectChains(cid int) []PageReport {
	query := `
		SELECT
			a.id,
			a.url,
			a.title
		FROM pagereports AS a
		LEFT JOIN pagereports AS b ON a.redirect_hash = b.url_hash
		WHERE a.redirect_hash != "" AND b.redirect_hash  != "" AND a.crawl_id = ? AND b.crawl_id = ?`

	return ds.pageReportsQuery(query, cid, cid)
}

func (ds *datastore) tooManyLinks(cid int) []PageReport {
	query := `
		SELECT
			pagereports.id,
			pagereports.url,
			pagereports.title
		FROM pagereports
		INNER JOIN (
			SELECT
				pagereport_id,
				count(distinct url_hash) as l
				FROM links
				WHERE crawl_id = ?
				GROUP BY pagereport_id
		) AS b ON pagereports.id = b.pagereport_id
		WHERE pagereports.crawl_id = ? and l > 100
	`

	return ds.pageReportsQuery(query, cid, cid)
}

func (ds *datastore) internalNoFollowLinks(cid int) []PageReport {
	query := `
		SELECT pagereports.id, pagereports.url, pagereports.title
		FROM pagereports 
		INNER JOIN (
			SELECT DISTINCT pagereport_id FROM links
			WHERE nofollow = 1 AND crawl_id = ?
		) AS b ON b.pagereport_id = pagereports.id
		WHERE pagereports.crawl_id = ?`

	return ds.pageReportsQuery(query, cid, cid)
}

func (ds *datastore) findSitemapPageReports(cid int) []PageReport {
	query := `
		SELECT pagereports.id, pagereports.url, pagereports.title
		FROM pagereports
		WHERE media_type = "text/html" AND status_code >= 200 AND status_code < 300
		AND (canonical IS NULL OR canonical = "" OR canonical = url) AND pagereports.crawl_id = ?`

	return ds.pageReportsQuery(query, cid)
}

func (ds *datastore) findExternalLinkWitoutNoFollow(cid int) []PageReport {
	query := `
		SELECT
			pagereports.id,
			pagereports.url,
			pagereports.title
		FROM pagereports
		INNER JOIN external_links ON pagereports.id = external_links.pagereport_id
		WHERE external_links.nofollow = 0 AND pagereports.crawl_id = ?
		GROUP BY pagereports.id`

	return ds.pageReportsQuery(query, cid)
}

func (ds *datastore) findCanonicalizedToNonCanonical(cid int) []PageReport {
	query := `
		SELECT
			a.id,
			a.url,
			a.title
		FROM pagereports AS a
		INNER JOIN pagereports AS b ON a.url = b.canonical
		WHERE a.crawl_id = ? AND b.crawl_id = ? AND a.canonical != "" AND a.canonical != a.url`

	return ds.pageReportsQuery(query, cid, cid)
}

func (ds *datastore) findRedirectLoops(cid int) []PageReport {
	query := `
		SELECT
			a.id,
			a.url,
			a.title
		FROM pagereports AS a
		INNER JOIN pagereports AS b ON a.redirect_hash = b.url_hash AND b.redirect_hash = a.url_hash
		WHERE a.crawl_id = ? AND b.crawl_id = ?`

	return ds.pageReportsQuery(query, cid, cid)
}

func (ds *datastore) findNotValidHeadingsOrder(cid int) []PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE crawl_id = ? AND valid_headings = 0`

	return ds.pageReportsQuery(query, cid)
}

func (ds *datastore) pageReportsQuery(query string, args ...interface{}) []PageReport {
	var pageReports []PageReport
	rows, err := ds.db.Query(query, args...)
	if err != nil {
		log.Println(err)
	}

	for rows.Next() {
		p := PageReport{}
		err := rows.Scan(&p.Id, &p.URL, &p.Title)
		if err != nil {
			log.Println(err)
			continue
		}

		pageReports = append(pageReports, p)
	}

	return pageReports
}
