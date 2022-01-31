package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("mysql", "root:root@tcp(0.0.0.0:6306)/seo?parseTime=true")
	if err != nil {
		panic(err)
	}

	if err := db.Ping(); err != nil {
		log.Printf("Unable to reach database: %v\n", err)
	}
}

func CountCrawled(cid int) int {
	row := db.QueryRow("SELECT count(*) FROM pagereports WHERE crawl_id = ?", cid)
	var c int
	if err := row.Scan(&c); err != nil {
		log.Printf("CountCrawled: %v\n", err)
	}

	return c
}

func CountByMediaType(cid int) map[string]int {
	m := make(map[string]int)

	rows, err := db.Query("SELECT media_type, count(*) FROM pagereports WHERE crawl_id = ? GROUP BY media_type", cid)
	if err != nil {
		log.Printf("CountByMediaType: %v\n", err)
		return m
	}

	for rows.Next() {
		var i string
		var v int
		err := rows.Scan(&i, &v)
		if err != nil {
			log.Println(err)
			continue
		}
		m[i] = v
	}
	return m
}

func userSignup(user, password string) {
	query := `INSERT INTO users (email, password) VALUES (?, ?)`
	stmt, _ := db.Prepare(query)
	defer stmt.Close()

	_, err := stmt.Exec(user, password)
	if err != nil {
		log.Printf("userSignup: %v\n", err)
	}
}

func findUserByEmail(email string) *User {
	u := User{}
	query := `SELECT id, email, password FROM users WHERE email = ?`

	row := db.QueryRow(query, email)
	err := row.Scan(&u.Id, &u.Email, &u.Password)
	if err != nil {
		log.Println(err)
		return &u
	}

	return &u
}

func findUserById(id int) *User {
	u := User{}
	query := `SELECT id, email, password FROM users WHERE id = ?`

	row := db.QueryRow(query, id)
	err := row.Scan(&u.Id, &u.Email, &u.Password)
	if err != nil {
		log.Println(err)
		return &u
	}

	return &u
}

func findCrawlUserId(cid int) (*User, error) {
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

	row := db.QueryRow(query, cid)
	err := row.Scan(&u.Id, &u.Email, &u.Password)
	if err != nil {
		log.Println(err)
		return &u, err
	}

	return &u, nil
}

func CountByStatusCode(cid int) map[int]int {
	m := make(map[int]int)
	query := `
		SELECT
			status_code,
			count(*)
		FROM pagereports
		WHERE crawl_id = ?
		GROUP BY status_code`

	rows, err := db.Query(query, cid)
	if err != nil {
		log.Println(err)
		return m
	}

	for rows.Next() {
		var i int
		var v int
		err := rows.Scan(&i, &v)
		if err != nil {
			log.Println(err)
			continue
		}
		m[i] = v
	}
	return m
}

func saveCrawl(p Project) int64 {
	stmt, _ := db.Prepare("INSERT INTO crawls (project_id) VALUES (?)")
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

func saveEndCrawl(cid int64, t time.Time) {
	stmt, _ := db.Prepare("UPDATE crawls SET end = ? WHERE id = ?")
	defer stmt.Close()
	_, err := stmt.Exec(t, cid)
	if err != nil {
		log.Printf("saveEndCrawl: %v\n", err)
	}
}

func getLastCrawl(p *Project) Crawl {
	row := db.QueryRow("SELECT id, start, end FROM crawls WHERE project_id = ? ORDER BY start DESC LIMIT 1", p.Id)

	crawl := Crawl{}
	err := row.Scan(&crawl.Id, &crawl.Start, &crawl.End)
	if err != nil {
		log.Printf("getLastCrawl: %v\n", err)
	}

	return crawl
}

func saveProject(s string, uid int) {
	stmt, _ := db.Prepare("INSERT INTO projects (url, user_id) VALUES (?, ?)")
	defer stmt.Close()
	_, err := stmt.Exec(s, uid)
	if err != nil {
		log.Printf("saveProject: %v\n", err)
	}
}

func findProjectsByUser(uid int) []Project {
	var projects []Project
	rows, err := db.Query("SELECT id, url, created FROM projects WHERE user_id = ?", uid)
	if err != nil {
		log.Println(err)
		return projects
	}

	for rows.Next() {
		p := Project{}
		err := rows.Scan(&p.Id, &p.URL, &p.Created)
		if err != nil {
			log.Println(err)
			continue
		}

		projects = append(projects, p)
	}

	return projects
}

func findCrawlById(cid int) Crawl {
	row := db.QueryRow("SELECT id, project_id, start, end FROM crawls WHERE id = ?", cid)

	c := Crawl{}
	err := row.Scan(&c.Id, &c.ProjectId, &c.Start, &c.End)
	if err != nil {
		log.Println(err)
		return c
	}

	return c
}

func findProjectById(id int, uid int) (Project, error) {
	row := db.QueryRow("SELECT id, url, created FROM projects WHERE id = ? AND user_id = ?", id, uid)

	p := Project{}
	err := row.Scan(&p.Id, &p.URL, &p.Created)
	if err != nil {
		log.Println(err)
		return p, err
	}

	return p, nil
}

func saveIssues(issues []Issue, cid int) {
	query := `
		INSERT INTO issues (pagereport_id, crawl_id, error_type)
		VALUES (?, ?, ?)`

	stmt, _ := db.Prepare(query)
	defer stmt.Close()

	for _, i := range issues {
		_, err := stmt.Exec(i.PageReportId, cid, i.ErrorType)
		if err != nil {
			log.Printf("saveIssues -> ID: %d ERROR: %s CRAWL: %d %v\n", i.PageReportId, i.ErrorType, cid, err)
			continue
		}
	}
}

func findIssues(cid int) map[string]IssueGroup {
	issues := map[string]IssueGroup{}
	query := `
		SELECT
			error_type,
			count(*)
		FROM issues
		WHERE crawl_id = ? GROUP BY error_type`

	rows, err := db.Query(query, cid)
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

func countIssuesByCrawl(cid int) int {
	var c int
	row := db.QueryRow("SELECT count(*) FROM issues WHERE crawl_id = ?", cid)
	if err := row.Scan(&c); err != nil {
		log.Println(err)
	}

	return c
}

func findErrorTypesByPage(pid, cid int) []string {
	var et []string
	query := `SELECT error_type FROM issues WHERE pagereport_id = ? and crawl_id = ? GROUP BY error_type`
	rows, err := db.Query(query, pid, cid)
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

func savePageReport(r *PageReport, cid int64) {
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
			size
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	stmt, err := db.Prepare(query)
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
		sqlString := "INSERT INTO links (pagereport_id, crawl_id, url, scheme, rel, text, external, url_hash) values "
		v := []interface{}{}
		for _, l := range r.Links {
			hash := hash(l.URL)
			sqlString += "(?, ?, ?, ?, ?, ?, ?, ?),"
			v = append(v, lid, cid, l.URL, l.parsedUrl.Scheme, l.Rel, l.Text, l.External, hash)
		}
		sqlString = sqlString[0 : len(sqlString)-1]
		stmt, err := db.Prepare(sqlString)
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
		sqlString := "INSERT INTO links (pagereport_id, crawl_id, url,scheme,  rel, text, external) values "
		v := []interface{}{}
		for _, l := range r.ExternalLinks {
			sqlString += "(?, ?, ?, ?, ?, ?, ?),"
			v = append(v, lid, cid, l.URL, l.parsedUrl.Scheme, l.Rel, l.Text, l.External)
		}
		sqlString = sqlString[0 : len(sqlString)-1]
		stmt, err := db.Prepare(sqlString)
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
		stmt, _ := db.Prepare(sqlString)
		defer stmt.Close()

		_, err := stmt.Exec(v...)
		if err != nil {
			log.Printf("savePageReport\nCID: %v\n Hreflangs: %+v\nError: %+v\n", cid, v, err)
		}
	}

	if len(r.Images) > 0 {
		sqlString := "INSERT INTO images (pagereport_id, url, alt) values "
		v := []interface{}{}
		for _, i := range r.Images {
			sqlString += "(?, ?, ?),"
			v = append(v, lid, i.URL, i.Alt)
		}
		sqlString = sqlString[0 : len(sqlString)-1]
		stmt, _ = db.Prepare(sqlString)
		defer stmt.Close()

		_, err := stmt.Exec(v...)
		if err != nil {
			log.Printf("savePageReport\nCID: %v\n Images: %+v\nError: %+v\n", cid, v, err)
		}
	}

	if len(r.Scripts) > 0 {
		sqlString := "INSERT INTO scripts (pagereport_id, url) values "
		v := []interface{}{}
		for _, s := range r.Scripts {
			sqlString += "(?, ?),"
			v = append(v, lid, s)
		}
		sqlString = sqlString[0 : len(sqlString)-1]
		stmt, _ := db.Prepare(sqlString)
		defer stmt.Close()

		_, err := stmt.Exec(v...)
		if err != nil {
			log.Printf("savePageReport\nCID: %v\n Scripts: %+v\nError: %+v\n", cid, v, err)
		}
	}

	if len(r.Styles) > 0 {
		sqlString := "INSERT INTO styles (pagereport_id, url) values "
		v := []interface{}{}

		for _, s := range r.Styles {
			sqlString += "(?, ?),"
			v = append(v, lid, s)

		}
		sqlString = sqlString[0 : len(sqlString)-1]
		stmt, _ := db.Prepare(sqlString)
		defer stmt.Close()

		_, err := stmt.Exec(v...)
		if err != nil {
			log.Printf("savePageReport\nCID: %v\n Styles: %+v\nError: %+v\n", cid, v, err)
		}
	}
}

func FindPageReportById(rid int) PageReport {
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
			size
		FROM pagereports
		WHERE id = ?`

	row := db.QueryRow(query, rid)

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
	)
	if err != nil {
		log.Println(err)
	}

	lrows, err := db.Query("SELECT url, rel, text, external FROM links WHERE external = false AND pagereport_id = ?", rid)
	if err != nil {
		log.Println(err)
	}

	for lrows.Next() {
		l := Link{}
		err = lrows.Scan(&l.URL, &l.Rel, &l.Text, &l.External)
		if err != nil {
			log.Println(err)
			continue
		}

		p.Links = append(p.Links, l)
	}

	lrows, err = db.Query("SELECT url, rel, text, external FROM links WHERE external = true AND pagereport_id = ?", rid)
	if err != nil {
		log.Println(err)
	}

	for lrows.Next() {
		l := Link{}
		err = lrows.Scan(&l.URL, &l.Rel, &l.Text, &l.External)
		if err != nil {
			log.Println(err)
			continue
		}

		p.ExternalLinks = append(p.ExternalLinks, l)
	}

	hrows, err := db.Query("SELECT to_url, to_lang FROM hreflangs WHERE pagereport_id = ?", rid)
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

	irows, err := db.Query("SELECT url, alt FROM images WHERE pagereport_id = ?", rid)
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

	scrows, err := db.Query("SELECT url FROM scripts WHERE pagereport_id = ?", rid)
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

	strows, err := db.Query("SELECT url FROM styles WHERE pagereport_id = ?", rid)
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

func FindPageReportsRedirectingToURL(u string, cid int) []PageReport {
	uh := hash(u)
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE redirect_hash = ? AND crawl_id = ?`

	return pageReportsQuery(query, uh, cid)
}

func FindPageReportsWithEmptyTitle(cid int) []PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE (title = "" OR title IS NULL) AND media_type = "text/html"
		AND status_code >=200 AND status_code < 300 AND crawl_id = ?`

	return pageReportsQuery(query, cid)
}

func Find40xPageReports(cid int) []PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE status_code >= 400 AND status_code < 500 AND crawl_id = ?`

	return pageReportsQuery(query, cid)
}

func Find30xPageReports(cid int) []PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE status_code >= 300 AND status_code < 400 AND crawl_id = ?`

	return pageReportsQuery(query, cid)
}

func Find50xPageReports(cid int) []PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE status_code >= 500 AND crawl_id = ?`

	return pageReportsQuery(query, cid)
}

func FindPageReportsWithLittleContent(cid int) []PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE words < 200 AND status_code >= 200 AND status_code < 300 AND media_type = "text/html" AND crawl_id = ?`

	return pageReportsQuery(query, cid)
}

func FindPageReportsWithShortTitle(cid int) []PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE length(title) > 0 AND length(title) < 20 AND media_type = "text/html" AND crawl_id = ?`

	return pageReportsQuery(query, cid)
}

func FindPageReportsWithLongTitle(cid int) []PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE length(title) > 60 AND media_type = "text/html" AND crawl_id = ?`

	return pageReportsQuery(query, cid)
}

func FindPageReportsWithDuplicatedTitle(cid int) []PageReport {
	query := `
		SELECT
			y.id,
			y.url,
			y.title
		FROM pagereports y
		INNER JOIN (
			SELECT
				title,
				count(*) AS c
			FROM pagereports
			WHERE crawl_id = ? AND status_code >= 200 AND status_code < 300 AND (canonical = "" OR canonical = url)
			GROUP BY title
			HAVING c > 1
		) d 
		ON d.title = y.title
		WHERE media_type = "text/html" AND length(y.title) > 0 AND crawl_id = ?`

	return pageReportsQuery(query, cid, cid)
}

func FindPageReportsWithoutH1(cid int) []PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE (h1 = "" OR h1 IS NULL) AND media_type = "text/html"
		AND status_code >= 200 AND status_code < 300 AND crawl_id = ?`

	return pageReportsQuery(query, cid)
}

func FindPageReportsWithEmptyDescription(cid int) []PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE (description = "" OR description IS NULL) AND media_type = "text/html"
		AND status_code >= 200 AND status_code < 300 AND crawl_id = ?`

	return pageReportsQuery(query, cid)
}

func FindPageReportsWithShortDescription(cid int) []PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE length(description) > 0 AND length(description) < 80 AND media_type = "text/html" AND crawl_id = ?`

	return pageReportsQuery(query, cid)
}

func FindPageReportsWithLongDescription(cid int) []PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE length(description) > 160 AND media_type = "text/html" AND crawl_id = ?`

	return pageReportsQuery(query, cid)
}

func FindPageReportsWithDuplicatedDescription(cid int) []PageReport {
	query := `
		SELECT
			y.id,
			y.url,
			y.title
		FROM pagereports y
		INNER JOIN (
			SELECT
				description,
				count(*) AS c
			FROM pagereports
			WHERE crawl_id = ? AND status_code >= 200 AND status_code < 300 AND (canonical = "" OR canonical = url)
			GROUP BY description
			HAVING c > 1
		) d 
		ON d.description = y.description
		WHERE y.media_type = "text/html" AND length(y.description) > 0 AND y.crawl_id = ?`

	return pageReportsQuery(query, cid, cid)
}

func FindImagesWithNoAlt(cid int) []PageReport {
	query := `
		SELECT
			pagereports.id,
			pagereports.url,
			pagereports.title
		FROM pagereports
		LEFT JOIN images ON images.pagereport_id = pagereports.id
		WHERE images.alt = "" AND pagereports.crawl_id = ?
		GROUP BY pagereports.id`

	return pageReportsQuery(query, cid)
}

func FindPageReportsWithNoLangAttr(cid int) []PageReport {
	query := `
		SELECT
			pagereports.id,
			pagereports.url,
			pagereports.title
		FROM pagereports
		WHERE (pagereports.lang = "" OR pagereports.lang = null) and media_type = "text/html" AND pagereports.crawl_id = ?`

	return pageReportsQuery(query, cid)
}

func FindPageReportsWithHTTPLinks(cid int) []PageReport {
	query := `
		SELECT
			pagereports.id,
			pagereports.url,
			pagereports.title
		FROM pagereports
		LEFT JOIN links ON links.pagereport_id = pagereports.id
		WHERE pagereports.scheme = "https" AND links.scheme = "http" AND pagereports.crawl_id = ? AND links.external = false
		GROUP BY links.pagereport_id
		HAVING count(links.pagereport_id) > 1`

	return pageReportsQuery(query, cid)
}

func FindMissingHrelangReturnLinks(cid int) []PageReport {
	query := `
		SELECT
			distinct pagereports.id,
			pagereports.URL,
			pagereports.Title
		FROM hreflangs
		LEFT JOIN hreflangs b ON hreflangs.crawl_id = b.crawl_id and hreflangs.from_hash = b.to_hash
		LEFT JOIN pagereports ON hreflangs.pagereport_id = pagereports.id
		WHERE  hreflangs.crawl_id = ? AND b.id IS NULL`

	return pageReportsQuery(query, cid)
}

func FindInLinks(s string, cid int) []PageReport {
	hash := hash(s)
	query := `
		SELECT 
			pagereports.id,
			pagereports.url,
			pagereports.Title
		FROM links
		LEFT JOIN pagereports ON pagereports.id = links.pagereport_id
		WHERE links.url_hash = ? AND pagereports.crawl_id = ?
		GROUP BY pagereports.id`

	return pageReportsQuery(query, hash, cid)
}

func findPageReportIssues(cid int, errorType string) []PageReport {
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
		)`

	return pageReportsQuery(query, errorType, cid)
}

func findRedirectChains(cid int) []PageReport {
	query := `
		SELECT
			a.id,
			a.url,
			a.title
		FROM pagereports AS a
		LEFT JOIN pagereports AS b ON a.redirect_hash = b.url_hash
		WHERE a.redirect_hash != "" AND b.redirect_hash  != "" AND a.crawl_id = ? AND b.crawl_id = ?`

	return pageReportsQuery(query, cid, cid)
}

func pageReportsQuery(query string, args ...interface{}) []PageReport {
	var pageReports []PageReport
	rows, err := db.Query(query, args...)
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
