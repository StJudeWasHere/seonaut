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

func savePageReport(r *PageReport, cid int64) {
	query := `
		INSERT INTO pagereports (
			crawl_id,
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
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	stmt, err := db.Prepare(query)
	if err != nil {
		log.Println(err)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(
		cid,
		r.URL,
		r.RedirectURL,
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
		sqlString := "INSERT INTO links (pagereport_id, url, rel, text, external) values "
		v := []interface{}{}
		for _, l := range r.Links {
			sqlString += "(?, ?, ?, ?, ?),"
			v = append(v, lid, l.URL, l.Rel, l.Text, l.External)
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
		sqlString := "INSERT INTO hreflangs (pagereport_id, url, lang) values "
		v := []interface{}{}
		for _, h := range r.Hreflangs {
			sqlString += "(?, ?, ?),"
			v = append(v, lid, h.URL, h.Lang)
		}
		sqlString = sqlString[0 : len(sqlString)-1]
		stmt, _ := db.Prepare(sqlString)
		defer stmt.Close()

		_, err := stmt.Exec(v...)
		if err != nil {
			log.Printf("Error in SavePageReport\nCID: %v\n Hreflangs: %+v\nError: %+v\n", cid, v, err)
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
			log.Printf("Error in SavePageReport\nCID: %v\n Images: %+v\nError: %+v\n", cid, v, err)
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
			log.Printf("Error in SavePageReport\nCID: %v\n Scripts: %+v\nError: %+v\n", cid, v, err)
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
			log.Printf("Error in SavePageReport\nCID: %v\n Styles: %+v\nError: %+v\n", cid, v, err)
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

	lrows, err := db.Query("SELECT url, rel, text, external FROM links WHERE pagereport_id = ?", rid)
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

	hrows, err := db.Query("SELECT url, lang FROM hreflangs WHERE pagereport_id = ?", rid)
	if err != nil {
		log.Println(err)
	}

	for hrows.Next() {
		h := Hreflang{}
		err = hrows.Scan(&h.URL, h.Lang)
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

func FindPageReportsWithEmptyTitle(cid int) []PageReport {
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
			size
		FROM pagereports
		WHERE (title = "" OR title IS NULL) AND media_type = "text/html" AND crawl_id = ?`

	rows, err := db.Query(query, cid)
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
		)
		if err != nil {
			log.Println(err)
			continue
		}

		pageReports = append(pageReports, p)
	}

	return pageReports
}

func Find40xPageReports(cid int) []PageReport {
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
			size
		FROM pagereports
		WHERE status_code >= 400 AND status_code < 500 AND crawl_id = ?`

	rows, err := db.Query(query, cid)
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
		)
		if err != nil {
			log.Println(err)
			continue
		}

		pageReports = append(pageReports, p)
	}

	return pageReports
}

func Find30xPageReports(cid int) []PageReport {
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
			size
		FROM pagereports
		WHERE status_code >= 300 AND status_code < 400 AND crawl_id = ?`

	rows, err := db.Query(query, cid)
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
		)
		if err != nil {
			log.Println(err)
			continue
		}

		pageReports = append(pageReports, p)
	}

	return pageReports
}

func Find50xPageReports(cid int) []PageReport {
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
			size
		FROM pagereports
		WHERE status_code >= 500 AND crawl_id = ?`

	rows, err := db.Query(query, cid)
	if err != nil {
		log.Println(err)
	}

	for rows.Next() {
		p := PageReport{}
		err := rows.Scan(
			&p.Id,
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
			continue
		}

		pageReports = append(pageReports, p)
	}

	return pageReports
}

func FindPageReportsWithLittleContent(cid int) []PageReport {
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
			size
		FROM pagereports
		WHERE words < 200 AND status_code >= 200 AND status_code < 300 AND media_type = "text/html" AND crawl_id = ?`

	rows, err := db.Query(query, cid)
	if err != nil {
		log.Println(err)
	}

	for rows.Next() {
		p := PageReport{}
		err := rows.Scan(
			&p.Id,
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
			continue
		}

		pageReports = append(pageReports, p)
	}

	return pageReports
}

func FindPageReportsWithShortTitle(cid int) []PageReport {
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
			size
		FROM pagereports
		WHERE length(title) > 0 AND length(title) < 20 AND media_type = "text/html" AND crawl_id = ?`

	rows, err := db.Query(query, cid)
	if err != nil {
		log.Println(err)
	}

	for rows.Next() {
		p := PageReport{}
		err := rows.Scan(
			&p.Id,
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
			continue
		}

		pageReports = append(pageReports, p)
	}

	return pageReports
}

func FindPageReportsWithLongTitle(cid int) []PageReport {
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
			size 
		FROM pagereports
		WHERE length(title) > 60 AND media_type = "text/html" AND crawl_id = ?`

	rows, err := db.Query(query, cid)
	if err != nil {
		log.Println(err)
	}

	for rows.Next() {
		p := PageReport{}
		err := rows.Scan(
			&p.Id,
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
			continue
		}

		pageReports = append(pageReports, p)
	}

	return pageReports
}

func FindPageReportsWithDuplicatedTitle(cid int) []PageReport {
	var pageReports []PageReport
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

	rows, err := db.Query(query, cid, cid)
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

func FindPageReportsWithEmptyDescription(cid int) []PageReport {
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
			size
		FROM pagereports
		WHERE (description = "" OR description IS NULL) AND media_type = "text/html" AND crawl_id = ?`

	rows, err := db.Query(query, cid)
	if err != nil {
		log.Println(err)
	}

	for rows.Next() {
		p := PageReport{}
		err := rows.Scan(
			&p.Id,
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
			continue
		}

		pageReports = append(pageReports, p)
	}

	return pageReports
}

func FindPageReportsWithShortDescription(cid int) []PageReport {
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
			size
		FROM pagereports
		WHERE length(description) > 0 AND length(description) < 80 AND media_type = "text/html" AND crawl_id = ?`

	rows, err := db.Query(query, cid)
	if err != nil {
		log.Println(err)
	}

	for rows.Next() {
		p := PageReport{}
		err := rows.Scan(
			&p.Id,
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
			continue
		}

		pageReports = append(pageReports, p)
	}

	return pageReports
}

func FindPageReportsWithLongDescription(cid int) []PageReport {
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
			size
		FROM pagereports
		WHERE length(description) > 160 AND media_type = "text/html" AND crawl_id = ?`

	rows, err := db.Query(query, cid)
	if err != nil {
		log.Println(err)
	}

	for rows.Next() {
		p := PageReport{}
		err := rows.Scan(
			&p.Id,
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
			continue
		}

		pageReports = append(pageReports, p)
	}

	return pageReports
}

func FindPageReportsWithDuplicatedDescription(cid int) []PageReport {
	var pageReports []PageReport
	query := `
		SELECT
			y.id,
			y.url,
			y.description
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
		WHERE media_type = "text/html" AND length(y.description) > 0 AND crawl_id = ?`

	rows, err := db.Query(query, cid, cid)
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

func CountCrawled(cid int) int {
	row := db.QueryRow("SELECT count(*) FROM pagereports WHERE crawl_id = ?", cid)
	var c int
	if err := row.Scan(&c); err != nil {
		log.Println(err)
	}

	return c
}

func CountByMediaType(cid int) map[string]int {
	m := make(map[string]int)

	rows, err := db.Query("SELECT media_type, count(*) FROM pagereports WHERE crawl_id = ? GROUP BY media_type", cid)
	if err != nil {
		log.Println(err)
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

func FindImagesWithNoAlt(cid int) []PageReport {
	pr := []PageReport{}
	query := `
		SELECT
			pagereports.id,
			pagereports.url,
			pagereports.title
		FROM pagereports
		LEFT JOIN images ON images.pagereport_id = pagereports.id
		WHERE images.alt = "" AND pagereports.crawl_id = ?
		GROUP BY pagereports.id`

	rows, err := db.Query(query, cid)
	if err != nil {
		log.Println(err)
		return pr
	}

	for rows.Next() {
		p := PageReport{}
		err := rows.Scan(&p.Id, &p.URL, &p.Title)
		if err != nil {
			log.Println(err)
			continue
		}

		pr = append(pr, p)
	}

	return pr
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
		log.Printf("Error in SaveCrawl\nProject: %+v\nError: %+v\n", p)
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
		log.Println(err)
	}
}

func getLastCrawl(p *Project) Crawl {
	row := db.QueryRow("SELECT id, start, end FROM crawls WHERE project_id = ? ORDER BY end DESC LIMIT 1", p.Id)

	crawl := Crawl{}
	err := row.Scan(&crawl.Id, &crawl.Start, &crawl.End)
	if err != nil {
		log.Println(err)
	}

	return crawl
}

func saveProject(s string) {
	stmt, _ := db.Prepare("INSERT INTO projects (url) VALUES (?)")
	defer stmt.Close()
	_, err := stmt.Exec(s)
	if err != nil {
		log.Println(err)
	}
}

func findProjects() []Project {
	var projects []Project
	rows, err := db.Query("SELECT id, url, created FROM projects")
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

func findProjectById(id int) Project {
	row := db.QueryRow("SELECT id, url, created FROM projects WHERE id = ?", id)

	p := Project{}
	err := row.Scan(&p.Id, &p.URL, &p.Created)
	if err != nil {
		log.Println(err)
	}

	return p
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
			log.Println(err)
			log.Printf("saveIssues -> ID: %s ERROR: %s CRAWL: %d\n", i.PageReportId, i.ErrorType, cid)
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

func findPageReportIssues(cid int, errorType string) []PageReport {
	pr := []PageReport{}
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

	rows, err := db.Query(query, errorType, cid)
	if err != nil {
		log.Println(err)
		return pr
	}

	for rows.Next() {
		p := PageReport{}
		err := rows.Scan(&p.Id, &p.URL, &p.Title)
		if err != nil {
			log.Println(err)
			continue
		}

		pr = append(pr, p)
	}

	return pr
}

func findRedirectChains(cid int) []PageReport {
	pr := []PageReport{}
	query := `
		SELECT
			a.id,
			a.url,
		FROM pagereports AS a
		LEFT JOIN pagereports AS b ON a.redirect_url = b.url
		WHERE a.redirect_url != "" AND b.redirect_url  != "" AND a.crawl_id = ? AND b.crawl_id = ?`

	rows, err := db.Query(query, cid)
	if err != nil {
		log.Println(err)
		return pr
	}

	for rows.Next() {
		p := PageReport{}
		err := rows.Scan(&p.Id, &p.URL)
		if err != nil {
			log.Println(err)
			continue
		}
	}

	return pr
}

func userSignup(user, password string) {
	query := `INSERT INTO users (email, password) VALUES (?, ?)`
	stmt, _ := db.Prepare(query)
	defer stmt.Close()

	_, err := stmt.Exec(user, password)
	if err != nil {
		log.Println(err)
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
