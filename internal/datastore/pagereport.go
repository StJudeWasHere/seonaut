package datastore

import (
	"log"

	"github.com/stjudewashere/seonaut/internal/crawler"
	"github.com/stjudewashere/seonaut/internal/helper"
)

func (ds *Datastore) SavePageReport(r *crawler.PageReport, cid int64) {
	urlHash := helper.Hash(r.URL)
	var redirectHash string
	if r.RedirectURL != "" {
		redirectHash = helper.Hash(r.RedirectURL)
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
			noindex,
			canonical,
			h1,
			h2,
			words,
			size,
			valid_headings,
			robotstxt_blocked,
			crawled
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

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
		r.Noindex,
		r.Canonical,
		r.H1,
		r.H2,
		r.Words,
		len(r.Body),
		r.ValidHeadings,
		r.BlockedByRobotstxt,
		r.Crawled,
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
			hash := helper.Hash(l.URL)
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
			v = append(v, lid, cid, r.Lang, h.URL, h.Lang, helper.Hash(r.URL), helper.Hash(h.URL))
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

func (ds *Datastore) FindAllPageReportsByCrawlId(cid int64) []crawler.PageReport {
	var pageReports []crawler.PageReport
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
			noindex,
			canonical,
			h1,
			h2,
			words,
			size,
			valid_headings,
			robotstxt_blocked,
			crawled
		FROM pagereports
		WHERE crawl_id = ?`

	rows, err := ds.db.Query(query, cid)
	if err != nil {
		log.Println(err)
	}

	for rows.Next() {
		p := crawler.PageReport{}
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
			&p.Noindex,
			&p.Canonical,
			&p.H1,
			&p.H2,
			&p.Words,
			&p.Size,
			&p.ValidHeadings,
			&p.BlockedByRobotstxt,
			&p.Crawled,
		)
		if err != nil {
			log.Println(err)
			continue
		}

		pageReports = append(pageReports, p)
	}

	return pageReports
}

func (ds *Datastore) FindAllPageReportsByCrawlIdAndErrorType(cid int64, et string) []crawler.PageReport {
	var pageReports []crawler.PageReport
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
			noindex,
			canonical,
			h1,
			h2,
			words,
			size,
			valid_headings,
			robotstxt_blocked,
			crawled
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
		p := crawler.PageReport{}
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
			&p.Noindex,
			&p.Canonical,
			&p.H1,
			&p.H2,
			&p.Words,
			&p.Size,
			&p.ValidHeadings,
			&p.BlockedByRobotstxt,
			&p.Crawled,
		)
		if err != nil {
			log.Println(err)
			continue
		}

		pageReports = append(pageReports, p)
	}

	return pageReports
}

func (ds *Datastore) FindPageReportById(rid int) crawler.PageReport {
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
			noindex,
			canonical,
			h1,
			h2,
			words,
			size,
			valid_headings,
			robotstxt_blocked,
			crawled
		FROM pagereports
		WHERE id = ?`

	row := ds.db.QueryRow(query, rid)

	p := crawler.PageReport{}
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
		&p.Noindex,
		&p.Canonical,
		&p.H1,
		&p.H2,
		&p.Words,
		&p.Size,
		&p.ValidHeadings,
		&p.BlockedByRobotstxt,
		&p.Crawled,
	)
	if err != nil {
		log.Println(err)
	}

	lrows, err := ds.db.Query("SELECT url, rel, nofollow, text FROM links WHERE pagereport_id = ? limit 25", rid)
	if err != nil {
		log.Println(err)
	}

	for lrows.Next() {
		l := crawler.Link{}
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
		l := crawler.Link{}
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
		h := crawler.Hreflang{}
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
		i := crawler.Image{}
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
