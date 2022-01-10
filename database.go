package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("mysql", "root:root@tcp(0.0.0.0:6306)/seo")
	if err != nil {
		panic(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("unable to reach database: %v", err)
	}
	fmt.Println("database is reachable")
}

func savePageReport(r *PageReport) {
	stmt, _ := db.Prepare("INSERT INTO pagereports (url, redirect_url, refresh, status_code, content_type, media_type, lang, title, description, robots, canonical, h1, h2, words, size) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	res, err := stmt.Exec(
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
		log.Fatal(err)
		return
	}

	lid, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
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
		stmt, _ = db.Prepare(sqlString)

		_, err := stmt.Exec(v...)
		if err != nil {
			log.Fatal(err)
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
		stmt, _ = db.Prepare(sqlString)

		_, err := stmt.Exec(v...)
		if err != nil {
			log.Fatal(err)
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

		_, err := stmt.Exec(v...)
		if err != nil {
			log.Fatal(err)
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
		stmt, _ = db.Prepare(sqlString)

		_, err := stmt.Exec(v...)
		if err != nil {
			log.Fatal(err)
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
		stmt, _ = db.Prepare(sqlString)

		_, err := stmt.Exec(v...)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func FindPageReports() []PageReport {
	var pageReports []PageReport

	sqlStr := "SELECT id, url, redirect_url, refresh, status_code, content_type, media_type, lang, title, description, robots, canonical, h1, h2, words, size FROM pagereports"
	rows, err := db.Query(sqlStr)
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		p := PageReport{}
		var pid int
		err := rows.Scan(&pid, &p.URL, &p.RedirectURL, &p.Refresh, &p.StatusCode, &p.ContentType, &p.MediaType, &p.Lang, &p.Title, &p.Description, &p.Robots, &p.Canonical, &p.H1, &p.H2, &p.Words, &p.Size)
		if err != nil {
			log.Fatal(err)
		}

		lrows, err := db.Query("SELECT url, rel, text, external FROM links WHERE pagereport_id = ?", pid)
		if err != nil {
			log.Fatal(err)
		}

		for lrows.Next() {
			l := Link{}
			err = lrows.Scan(&l.URL, &l.Rel, &l.Text, &l.External)
			if err != nil {
				log.Fatal(err)
			}
			p.Links = append(p.Links, l)
		}

		hrows, err := db.Query("SELECT url, lang FROM hreflangs WHERE pagereport_id = ?", pid)
		if err != nil {
			log.Fatal(err)
		}

		for hrows.Next() {
			h := Hreflang{}
			err = hrows.Scan(&h.URL, h.Lang)
			if err != nil {
				log.Fatal(err)
			}
			p.Hreflangs = append(p.Hreflangs, h)
		}

		irows, err := db.Query("SELECT url, alt FROM images WHERE pagereport_id = ?", pid)
		if err != nil {
			log.Fatal(err)
		}
		for irows.Next() {
			i := Image{}
			err = irows.Scan(&i.URL, &i.Alt)
			if err != nil {
				log.Fatal(err)
			}

			p.Images = append(p.Images, i)
		}

		scrows, err := db.Query("SELECT url FROM scripts WHERE pagereport_id = ?", pid)
		if err != nil {
			log.Fatal(err)
		}
		for scrows.Next() {
			var url string
			err = scrows.Scan(&url)
			if err != nil {
				log.Fatal(err)
			}

			p.Scripts = append(p.Scripts, url)
		}

		strows, err := db.Query("SELECT url FROM styles WHERE pagereport_id = ?", pid)
		if err != nil {
			log.Fatal(err)
		}
		for strows.Next() {
			var url string
			err = strows.Scan(&url)
			if err != nil {
				log.Fatal(err)
			}

			p.Styles = append(p.Styles, url)
		}

		pageReports = append(pageReports, p)
	}

	return pageReports
}

func FindPageReportsWithEmptyTitle() []PageReport {
	var pageReports []PageReport

	sqlStr := "SELECT id, url, redirect_url, refresh, status_code, content_type, media_type, lang, title, description, robots, canonical, h1, h2, words, size FROM pagereports WHERE (title = \"\" OR title is NULL) AND media_type = \"text/html\""
	rows, err := db.Query(sqlStr)
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		p := PageReport{}
		var pid int
		err := rows.Scan(&pid, &p.URL, &p.RedirectURL, &p.Refresh, &p.StatusCode, &p.ContentType, &p.MediaType, &p.Lang, &p.Title, &p.Description, &p.Robots, &p.Canonical, &p.H1, &p.H2, &p.Words, &p.Size)
		if err != nil {
			log.Fatal(err)
		}

		pageReports = append(pageReports, p)
	}

	return pageReports
}

func FindPageReportsWithShortTitle() []PageReport {
	var pageReports []PageReport

	sqlStr := "SELECT id, url, redirect_url, refresh, status_code, content_type, media_type, lang, title, description, robots, canonical, h1, h2, words, size FROM pagereports WHERE length(title) > 0 AND length(title) < 20 AND media_type = \"text/html\""
	rows, err := db.Query(sqlStr)
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		p := PageReport{}
		var pid int
		err := rows.Scan(&pid, &p.URL, &p.RedirectURL, &p.Refresh, &p.StatusCode, &p.ContentType, &p.MediaType, &p.Lang, &p.Title, &p.Description, &p.Robots, &p.Canonical, &p.H1, &p.H2, &p.Words, &p.Size)
		if err != nil {
			log.Fatal(err)
		}

		pageReports = append(pageReports, p)
	}

	return pageReports
}

func FindPageReportsWithLongTitle() []PageReport {
	var pageReports []PageReport

	sqlStr := "SELECT id, url, redirect_url, refresh, status_code, content_type, media_type, lang, title, description, robots, canonical, h1, h2, words, size FROM pagereports WHERE length(title) > 60 AND media_type = \"text/html\""
	rows, err := db.Query(sqlStr)
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		p := PageReport{}
		var pid int
		err := rows.Scan(&pid, &p.URL, &p.RedirectURL, &p.Refresh, &p.StatusCode, &p.ContentType, &p.MediaType, &p.Lang, &p.Title, &p.Description, &p.Robots, &p.Canonical, &p.H1, &p.H2, &p.Words, &p.Size)
		if err != nil {
			log.Fatal(err)
		}

		pageReports = append(pageReports, p)
	}

	return pageReports
}

func FindPageReportsWithDuplicatedTitle() []PageReport {
	var pageReports []PageReport
	sqlStr := "select y.id, y.url, y.title from pagereports y inner join (select title, count(*) as c from pagereports group by title having c > 1) d on d.title = y.title where media_type = \"text/html\" AND length(y.title) > 0"
	rows, err := db.Query(sqlStr)
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		p := PageReport{}
		var pid int
		err := rows.Scan(&pid, &p.URL, &p.Title)
		if err != nil {
			log.Fatal(err)
		}

		pageReports = append(pageReports, p)
	}

	return pageReports
}

func FindPageReportsWithEmptyDescription() []PageReport {
	var pageReports []PageReport

	sqlStr := "SELECT id, url, redirect_url, refresh, status_code, content_type, media_type, lang, title, description, robots, canonical, h1, h2, words, size FROM pagereports WHERE (description = \"\" OR description is NULL) AND media_type = \"text/html\""
	rows, err := db.Query(sqlStr)
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		p := PageReport{}
		var pid int
		err := rows.Scan(&pid, &p.URL, &p.RedirectURL, &p.Refresh, &p.StatusCode, &p.ContentType, &p.MediaType, &p.Lang, &p.Title, &p.Description, &p.Robots, &p.Canonical, &p.H1, &p.H2, &p.Words, &p.Size)
		if err != nil {
			log.Fatal(err)
		}

		pageReports = append(pageReports, p)
	}

	return pageReports
}

func FindPageReportsWithShortDescription() []PageReport {
	var pageReports []PageReport

	sqlStr := "SELECT id, url, redirect_url, refresh, status_code, content_type, media_type, lang, title, description, robots, canonical, h1, h2, words, size FROM pagereports WHERE length(description) > 0 AND length(description) < 80 AND media_type = \"text/html\""
	rows, err := db.Query(sqlStr)
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		p := PageReport{}
		var pid int
		err := rows.Scan(&pid, &p.URL, &p.RedirectURL, &p.Refresh, &p.StatusCode, &p.ContentType, &p.MediaType, &p.Lang, &p.Title, &p.Description, &p.Robots, &p.Canonical, &p.H1, &p.H2, &p.Words, &p.Size)
		if err != nil {
			log.Fatal(err)
		}

		pageReports = append(pageReports, p)
	}

	return pageReports
}

func FindPageReportsWithLongDescription() []PageReport {
	var pageReports []PageReport

	sqlStr := "SELECT id, url, redirect_url, refresh, status_code, content_type, media_type, lang, title, description, robots, canonical, h1, h2, words, size FROM pagereports WHERE length(description) > 160 AND media_type = \"text/html\""
	rows, err := db.Query(sqlStr)
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		p := PageReport{}
		var pid int
		err := rows.Scan(&pid, &p.URL, &p.RedirectURL, &p.Refresh, &p.StatusCode, &p.ContentType, &p.MediaType, &p.Lang, &p.Title, &p.Description, &p.Robots, &p.Canonical, &p.H1, &p.H2, &p.Words, &p.Size)
		if err != nil {
			log.Fatal(err)
		}

		pageReports = append(pageReports, p)
	}

	return pageReports
}

func FindPageReportsWithDuplicatedDescription() []PageReport {
	var pageReports []PageReport
	sqlStr := "select y.id, y.url, y.description from pagereports y inner join (select description, count(*) as c from pagereports group by description having c > 1) d on d.description = y.description where media_type = \"text/html\" AND length(y.description) > 0"
	rows, err := db.Query(sqlStr)
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		p := PageReport{}
		var pid int
		err := rows.Scan(&pid, &p.URL, &p.Title)
		if err != nil {
			log.Fatal(err)
		}

		pageReports = append(pageReports, p)
	}

	return pageReports
}

func CountCrawled() int {
	row := db.QueryRow("SELECT count(*) FROM pagereports")
	var c int
	if err := row.Scan(&c); err != nil {
		log.Fatal(err)
	}

	return c
}

func CountByMediaType() map[string]int {
	m := make(map[string]int)

	rows, err := db.Query("select media_type, count(*) from pagereports group by media_type")
	if err != nil {
		log.Fatal(err)
		return m
	}

	for rows.Next() {
		var i string
		var v int
		err := rows.Scan(&i, &v)
		if err != nil {
			log.Fatal(err)
			continue
		}
		m[i] = v
	}
	return m
}

func CountByStatusCode() map[int]int {
	m := make(map[int]int)

	rows, err := db.Query("select status_code, count(*) from pagereports group by status_code")
	if err != nil {
		log.Fatal(err)
		return m
	}

	for rows.Next() {
		var i int
		var v int
		err := rows.Scan(&i, &v)
		if err != nil {
			log.Fatal(err)
			continue
		}
		m[i] = v
	}
	return m
}
