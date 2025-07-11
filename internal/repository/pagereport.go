package repository

import (
	"database/sql"
	"log"
	"math"
	"net/url"

	"github.com/stjudewashere/seonaut/internal/models"
)

type PageReportRepository struct {
	DB *sql.DB
}

// Save a pagereport and all its related data to the database. Once the pagereport is saved it uses its id to
// save the associated links, external_links, images, scripts and so on.
func (ds *PageReportRepository) SavePageReport(r *models.PageReport, cid int64) (*models.PageReport, error) {
	urlHash := Hash(r.URL)
	var redirectHash string
	if r.RedirectURL != "" {
		redirectHash = Hash(r.RedirectURL)
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
			robotstxt_blocked,
			crawled,
			in_sitemap,
			depth,
			body_hash,
			ttfb
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	stmt, err := ds.DB.Prepare(query)
	if err != nil {
		return r, err
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
		Truncate(r.Title, 2048),
		Truncate(r.Description, 2048),
		r.Robots,
		r.Noindex,
		r.Canonical,
		Truncate(r.H1, 1024),
		Truncate(r.H2, 1024),
		r.Words,
		r.Size,
		r.BlockedByRobotstxt,
		r.Crawled,
		r.InSitemap,
		r.Depth,
		r.BodyHash,
		r.TTFB,
	)
	if err != nil {
		return r, err
	}

	r.Id, err = res.LastInsertId()
	if err != nil {
		return r, err
	}

	f := []func(*models.PageReport, int64) error{
		ds.SavePageReportLinks,
		ds.SavePageReportExternalLinks,
		ds.SavePageReportHreflangs,
		ds.SavePageReportImages,
		ds.SavePageReportIframes,
		ds.SavePageReportAudios,
		ds.SavePageReportVideos,
		ds.SavePageReportScripts,
		ds.SavePageReportStyles,
	}

	for _, sf := range f {
		err = sf(r, cid)
		if err != nil {
			log.Println(err)
		}
	}

	return r, nil
}

// Save pagereport internal links.
func (ds *PageReportRepository) SavePageReportLinks(r *models.PageReport, cid int64) error {
	if len(r.Links) == 0 {
		return nil
	}

	sqlString := "INSERT INTO links (pagereport_id, crawl_id, url, scheme, rel, nofollow, text, url_hash) values "
	v := []interface{}{}
	for _, l := range r.Links {
		hash := Hash(l.URL)
		sqlString += "(?, ?, ?, ?, ?, ?, ?, ?),"
		v = append(v, r.Id, cid, l.URL, l.ParsedURL.Scheme, l.Rel, l.NoFollow, Truncate(l.Text, 1024), hash)
	}
	sqlString = sqlString[0 : len(sqlString)-1]
	stmt, err := ds.DB.Prepare(sqlString)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(v...)
	return err
}

// Save pagereport external links.
func (ds *PageReportRepository) SavePageReportExternalLinks(r *models.PageReport, cid int64) error {
	if len(r.ExternalLinks) == 0 {
		return nil
	}

	sqlString := "INSERT INTO external_links (pagereport_id, crawl_id, url, rel, nofollow, text, sponsored, ugc, status_code) values "
	v := []interface{}{}
	for _, l := range r.ExternalLinks {
		sqlString += "(?, ?, ?, ?, ?, ?, ?, ?, ?),"
		v = append(v, r.Id, cid, l.URL, l.Rel, l.NoFollow, Truncate(l.Text, 1024), l.Sponsored, l.UGC, l.StatusCode)
	}
	sqlString = sqlString[0 : len(sqlString)-1]
	stmt, err := ds.DB.Prepare(sqlString)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(v...)
	return err
}

// Save pagereport hreflangs.
func (ds *PageReportRepository) SavePageReportHreflangs(r *models.PageReport, cid int64) error {
	if len(r.Hreflangs) == 0 {
		return nil
	}

	sqlString := "INSERT INTO hreflangs (pagereport_id, crawl_id, from_lang, to_url, to_lang, from_hash, to_hash) values "
	v := []interface{}{}
	for _, h := range r.Hreflangs {
		sqlString += "(?, ?, ?, ?, ?, ?, ?),"
		v = append(v, r.Id, cid, r.Lang, h.URL, h.Lang, Hash(r.URL), Hash(h.URL))
	}
	sqlString = sqlString[0 : len(sqlString)-1]
	stmt, _ := ds.DB.Prepare(sqlString)
	defer stmt.Close()

	_, err := stmt.Exec(v...)
	return err
}

// Save pagereport images.
func (ds *PageReportRepository) SavePageReportImages(r *models.PageReport, cid int64) error {
	if len(r.Images) == 0 {
		return nil
	}

	sqlString := "INSERT INTO images (pagereport_id, url, alt, crawl_id) values "
	v := []interface{}{}
	for _, i := range r.Images {
		sqlString += "(?, ?, ?, ?),"
		v = append(v, r.Id, i.URL, Truncate(i.Alt, 1024), cid)
	}
	sqlString = sqlString[0 : len(sqlString)-1]
	stmt, _ := ds.DB.Prepare(sqlString)
	defer stmt.Close()

	_, err := stmt.Exec(v...)
	return err
}

// Save pagereport iframes.
func (ds *PageReportRepository) SavePageReportIframes(r *models.PageReport, cid int64) error {
	if len(r.Iframes) == 0 {
		return nil
	}
	sqlString := "INSERT INTO iframes (pagereport_id, url, crawl_id) values "

	v := []interface{}{}
	for _, i := range r.Iframes {
		sqlString += "(?, ?, ?),"
		v = append(v, r.Id, i, cid)
	}
	sqlString = sqlString[0 : len(sqlString)-1]
	stmt, _ := ds.DB.Prepare(sqlString)
	defer stmt.Close()

	_, err := stmt.Exec(v...)
	return err
}

// Save pagereport audios.
func (ds *PageReportRepository) SavePageReportAudios(r *models.PageReport, cid int64) error {
	if len(r.Audios) == 0 {
		return nil
	}

	sqlString := "INSERT INTO audios (pagereport_id, url, crawl_id) values "

	v := []interface{}{}
	for _, i := range r.Audios {
		sqlString += "(?, ?, ?),"
		v = append(v, r.Id, i, cid)
	}
	sqlString = sqlString[0 : len(sqlString)-1]
	stmt, _ := ds.DB.Prepare(sqlString)
	defer stmt.Close()

	_, err := stmt.Exec(v...)
	return err
}

// Save pagereport videos.
func (ds *PageReportRepository) SavePageReportVideos(r *models.PageReport, cid int64) error {
	if len(r.Videos) == 0 {
		return nil
	}

	sqlString := "INSERT INTO videos (pagereport_id, url, poster, crawl_id) values "

	v := []interface{}{}
	for _, i := range r.Videos {
		sqlString += "(?, ?, ?, ?),"
		v = append(v, r.Id, i.URL, i.Poster, cid)
	}
	sqlString = sqlString[0 : len(sqlString)-1]
	stmt, _ := ds.DB.Prepare(sqlString)
	defer stmt.Close()

	_, err := stmt.Exec(v...)
	return err
}

// Save pagereport scripts.
func (ds *PageReportRepository) SavePageReportScripts(r *models.PageReport, cid int64) error {
	if len(r.Scripts) == 0 {
		return nil
	}

	sqlString := "INSERT INTO scripts (pagereport_id, url, crawl_id) values "
	v := []interface{}{}
	for _, s := range r.Scripts {
		sqlString += "(?, ?, ?),"
		v = append(v, r.Id, s, cid)
	}
	sqlString = sqlString[0 : len(sqlString)-1]
	stmt, _ := ds.DB.Prepare(sqlString)
	defer stmt.Close()

	_, err := stmt.Exec(v...)
	return err
}

// Save pagereport styles.
func (ds *PageReportRepository) SavePageReportStyles(r *models.PageReport, cid int64) error {
	if len(r.Styles) == 0 {
		return nil
	}

	sqlString := "INSERT INTO styles (pagereport_id, url, crawl_id) values "
	v := []interface{}{}

	for _, s := range r.Styles {
		sqlString += "(?, ?, ?),"
		v = append(v, r.Id, s, cid)

	}
	sqlString = sqlString[0 : len(sqlString)-1]
	stmt, _ := ds.DB.Prepare(sqlString)
	defer stmt.Close()

	_, err := stmt.Exec(v...)
	return err
}

// FindAllPageReportsByCrawlId returns a channel where it streams all the crawl's page reports.
// Once it is done it closes the channel.
func (ds *PageReportRepository) FindAllPageReportsByCrawlId(cid int64) <-chan *models.PageReport {
	prStream := make(chan *models.PageReport)

	go func() {
		defer close(prStream)

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
				robotstxt_blocked,
				crawled,
				in_sitemap,
				depth,
				body_hash,
				ttfb
			FROM pagereports
			WHERE crawl_id = ?`

		rows, err := ds.DB.Query(query, cid)
		if err != nil {
			log.Println(err)
		}

		for rows.Next() {
			p := &models.PageReport{}
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
				&p.BlockedByRobotstxt,
				&p.Crawled,
				&p.InSitemap,
				&p.Depth,
				&p.BodyHash,
				&p.TTFB,
			)
			if err != nil {
				log.Println(err)
				continue
			}

			prStream <- p
		}
	}()

	return prStream
}

// FindAllPageReportsByCrawlIdAndErrorType returns a channel of pagereports where it streams all the reports
// for the specified crawl and error type. Once it is done it closes the channel.
func (ds *PageReportRepository) FindAllPageReportsByCrawlIdAndErrorType(cid int64, et string) <-chan *models.PageReport {
	prStream := make(chan *models.PageReport)

	go func() {
		defer close(prStream)

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
				robotstxt_blocked,
				crawled,
				in_sitemap,
				depth,
				body_hash,
				ttfb
			FROM pagereports
			WHERE crawl_id = ?
			AND id IN (
				SELECT
					pagereport_id
				FROM issues
				INNER JOIN issue_types ON issue_types.id = issues.issue_type_id
				WHERE issue_types.type = ? AND crawl_id = ?
			)`

		rows, err := ds.DB.Query(query, cid, et, cid)
		if err != nil {
			log.Println(err)
		}

		for rows.Next() {
			p := &models.PageReport{}
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
				&p.BlockedByRobotstxt,
				&p.Crawled,
				&p.InSitemap,
				&p.Depth,
				&p.BodyHash,
				&p.TTFB,
			)
			if err != nil {
				log.Println(err)
				continue
			}

			prStream <- p
		}
	}()

	return prStream
}

// FindPageReportById returns a PageReport Model with all its hreflang tags
func (ds *PageReportRepository) FindPageReportById(rid int) models.PageReport {
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
			(robots LIKE '%nofollow%' OR robots LIKE '%none%') AS nofollow,
			canonical,
			h1,
			h2,
			words,
			size,
			robotstxt_blocked,
			crawled,
			in_sitemap,
			depth,
			body_hash,
			ttfb
		FROM pagereports
		WHERE id = ?`

	row := ds.DB.QueryRow(query, rid)

	p := models.PageReport{}
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
		&p.Nofollow,
		&p.Canonical,
		&p.H1,
		&p.H2,
		&p.Words,
		&p.Size,
		&p.BlockedByRobotstxt,
		&p.Crawled,
		&p.InSitemap,
		&p.Depth,
		&p.BodyHash,
		&p.TTFB,
	)
	if err != nil {
		log.Println(err)
	}

	p.ParsedURL, err = url.Parse(p.URL)
	if err != nil {
		log.Printf("error parsing url %s %v", p.URL, err)
	}

	return p
}

// Find images in an specific pagereport.
func (ds *PageReportRepository) FindPageReportHreflangs(pageReport *models.PageReport, cid int64) []models.Hreflang {
	hreflangs := []models.Hreflang{}

	hrows, err := ds.DB.Query("SELECT to_url, to_lang FROM hreflangs WHERE pagereport_id = ?", pageReport.Id)
	if err != nil {
		log.Println(err)
	}

	for hrows.Next() {
		h := models.Hreflang{}
		err = hrows.Scan(&h.URL, &h.Lang)
		if err != nil {
			log.Println(err)
			continue
		}

		hreflangs = append(hreflangs, h)
	}

	return hreflangs
}

// Find images in an specific pagereport.
func (ds *PageReportRepository) FindPageReportImages(pageReport *models.PageReport, cid int64) []models.Image {
	images := []models.Image{}

	irows, err := ds.DB.Query("SELECT url, alt FROM images WHERE pagereport_id = ?", pageReport.Id)
	if err != nil {
		log.Println(err)
	}

	for irows.Next() {
		i := models.Image{}
		err = irows.Scan(&i.URL, &i.Alt)
		if err != nil {
			log.Println(err)
			continue
		}

		images = append(images, i)
	}

	return images
}

// Find iframes in an specific pagereport.
func (ds *PageReportRepository) FindPageReportIframes(pageReport *models.PageReport, cid int64) []string {
	iframes := []string{}

	ifrows, err := ds.DB.Query("SELECT url FROM iframes WHERE pagereport_id = ?", pageReport.Id)
	if err != nil {
		log.Println(err)
	}

	for ifrows.Next() {
		var url string
		err = ifrows.Scan(&url)
		if err != nil {
			log.Println(err)
			continue
		}

		iframes = append(iframes, url)
	}

	return iframes
}

// Find audios in an specific pagereport.
func (ds *PageReportRepository) FindPageReportAudios(pageReport *models.PageReport, cid int64) []string {
	audios := []string{}

	arows, err := ds.DB.Query("SELECT url FROM audios WHERE pagereport_id = ?", pageReport.Id)
	if err != nil {
		log.Println(err)
	}

	for arows.Next() {
		var url string
		err = arows.Scan(&url)
		if err != nil {
			log.Println(err)
			continue
		}

		audios = append(audios, url)
	}

	return audios
}

// Find videos in an specific pagereport.
func (ds *PageReportRepository) FindPageReportVideos(pageReport *models.PageReport, cid int64) []models.Video {
	videos := []models.Video{}

	vrows, err := ds.DB.Query("SELECT url, poster FROM videos WHERE pagereport_id = ?", pageReport.Id)
	if err != nil {
		log.Println(err)
	}

	for vrows.Next() {
		var video models.Video
		err = vrows.Scan(&video.URL, &video.Poster)
		if err != nil {
			log.Println(err)
			continue
		}

		videos = append(videos, video)
	}

	return videos
}

// Find the scripts of an specific pagereport.
func (ds *PageReportRepository) FindPageReportScripts(pageReport *models.PageReport, cid int64) []string {
	scripts := []string{}

	scrows, err := ds.DB.Query("SELECT url FROM scripts WHERE pagereport_id = ?", pageReport.Id)
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

		scripts = append(scripts, url)
	}

	return scripts
}

// Find the styles of an specific pagereport
func (ds *PageReportRepository) FindPageReportStyles(pageReport *models.PageReport, cid int64) []string {
	styles := []string{}

	strows, err := ds.DB.Query("SELECT url FROM styles WHERE pagereport_id = ?", pageReport.Id)
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

		styles = append(styles, url)
	}

	return styles
}

// FindLinks returns a slice of paginated InternalLinks. The page is specified in the "p" parameter.
func (ds *PageReportRepository) FindLinks(pageReport *models.PageReport, cid int64, p int) []models.InternalLink {
	max := paginationMax
	offset := max * (p - 1)
	links := []models.InternalLink{}

	query := `
		SELECT
			pagereports.id,
			pagereports.url,
			pagereports.title,
			pagereports.crawled,
			links.url,
			links.rel,
			links.nofollow,
			links.text
		FROM links
		LEFT JOIN pagereports ON links.url_hash = pagereports.url_hash
		WHERE links.pagereport_id = ? and pagereports.crawl_id = ?
		LIMIT ?,?
	`

	lrows, err := ds.DB.Query(query, pageReport.Id, cid, offset, max)
	if err != nil {
		log.Println(err)
	}

	for lrows.Next() {
		l := models.InternalLink{}
		err = lrows.Scan(
			&l.PageReport.Id,
			&l.PageReport.URL,
			&l.PageReport.Title,
			&l.PageReport.Crawled,
			&l.Link.URL,
			&l.Link.Rel,
			&l.Link.NoFollow,
			&l.Link.Text,
		)
		if err != nil {
			log.Println(err)
			continue
		}

		links = append(links, l)
	}

	return links
}

// FindExternalLinks returns a slice of paginated Link models.
// The page to retrieve is specified in the "p" paramenter.
func (ds *PageReportRepository) FindExternalLinks(pageReport *models.PageReport, cid int64, p int) []models.Link {
	max := paginationMax
	offset := max * (p - 1)
	links := []models.Link{}

	query := `
		SELECT
			url,
			rel,
			nofollow,
			text,
			sponsored,
			ugc,
			status_code
		FROM external_links
		WHERE pagereport_id = ?
		LIMIT ?,?
	`

	lrows, err := ds.DB.Query(query, pageReport.Id, offset, max)
	if err != nil {
		log.Println(err)
	}

	for lrows.Next() {
		l := models.Link{}
		err = lrows.Scan(&l.URL, &l.Rel, &l.NoFollow, &l.Text, &l.Sponsored, &l.UGC, &l.StatusCode)
		if err != nil {
			log.Println(err)
			continue
		}

		links = append(links, l)
	}

	return links
}

// FindSitemapPageReports returns a channel of models.PageReport that is used to stream all
// the PageReports that are eligible to be added to a sitemap.xml file.
func (ds *PageReportRepository) FindSitemapPageReports(cid int64) <-chan *models.PageReport {
	prStream := make(chan *models.PageReport)

	go func() {
		defer close(prStream)

		query := `
			SELECT pagereports.id, pagereports.url, pagereports.title
			FROM pagereports
			WHERE media_type = "text/html" AND status_code >= 200 AND status_code < 300
			AND (canonical IS NULL OR canonical = "" OR canonical = url) AND pagereports.crawl_id = ?
			AND crawled = 1`

		rows, err := ds.DB.Query(query, cid)
		if err != nil {
			log.Println(err)
		}

		for rows.Next() {
			p := &models.PageReport{}
			err := rows.Scan(&p.Id, &p.URL, &p.Title)
			if err != nil {
				log.Println(err)
				continue
			}

			prStream <- p
		}
	}()

	return prStream
}

// FindPaginatedPageReports returns a paginated slice of models.PageReport.
// The page to be retrieved is specidied in the "p" parameter. This method also allows for
// "term" search in case it is not an empty string "".
func (ds *PageReportRepository) FindPaginatedPageReports(cid int64, p int, term string) []models.PageReport {
	max := paginationMax
	offset := max * (p - 1)
	args := []interface{}{term, cid}
	var pageReports []models.PageReport

	query := `
		SELECT
			id,
			url,
			title,
			(CASE WHEN url = ? THEN 1 ELSE 0 END) AS exact_match
		FROM pagereports
		WHERE crawl_id = ?
			AND crawled = 1`

	if term != "" {
		query += ` AND MATCH (url) AGAINST (? IN NATURAL LANGUAGE MODE)`
		args = append(args, term)
	}

	query += `
		ORDER BY exact_match DESC, url ASC
		LIMIT ?, ?`

	args = append(args, offset, max)

	rows, err := ds.DB.Query(query, args...)
	if err != nil {
		log.Println(err)
		return pageReports
	}

	for rows.Next() {
		var e bool
		p := models.PageReport{}
		err := rows.Scan(&p.Id, &p.URL, &p.Title, &e)
		if err != nil {
			log.Println(err)
			continue
		}

		pageReports = append(pageReports, p)
	}

	return pageReports
}

// GetNumberOfPagesForPageReport returns the total number of pageReport pages.
// This method can be used to build a paginator.
func (ds *PageReportRepository) GetNumberOfPagesForPageReport(cid int64, term string) int {
	query := `
		SELECT count(id)
		FROM pagereports
		WHERE crawl_id  = ?
			AND crawled = 1`

	args := []interface{}{cid}
	if term != "" {
		query += ` AND MATCH (url) AGAINST (? IN NATURAL LANGUAGE MODE)`
		args = append(args, term)
	}

	row := ds.DB.QueryRow(query, args...)
	var c int
	if err := row.Scan(&c); err != nil {
		log.Printf("GetNumberOfPagesForPageReport: %v\n", err)
	}
	var f float64 = float64(c) / float64(paginationMax)
	return int(math.Ceil(f))
}

// FindInLinks Returns a paginated slice of models.InternalLink models.
// The page number to be retrieved is specified in the "p" parameter.
func (ds *PageReportRepository) FindInLinks(s string, cid int64, p int) []models.InternalLink {
	max := paginationMax
	offset := max * (p - 1)

	hash := Hash(s)
	query := `
		SELECT
			pagereports.id,
			pagereports.url,
			pagereports.title,
			links.nofollow,
			links.text
		FROM links
		LEFT JOIN pagereports ON pagereports.id = links.pagereport_id
		WHERE links.url_hash = ? AND pagereports.crawl_id = ? AND pagereports.crawled = 1
		LIMIT ?,?`

	var internalLinks []models.InternalLink
	rows, err := ds.DB.Query(query, hash, cid, offset, max)
	if err != nil {
		log.Println(err)
	}

	for rows.Next() {
		il := models.InternalLink{}
		err := rows.Scan(&il.PageReport.Id, &il.PageReport.URL, &il.PageReport.Title, &il.Link.NoFollow, &il.Link.Text)
		if err != nil {
			log.Println(err)
			continue
		}

		internalLinks = append(internalLinks, il)
	}

	return internalLinks
}

// FindPageReportsRedirectingToURL returns a paginated slice of models.PageReport that are being redirected to
// a specidied URL. The page number is set in the "p" paramenter.
func (ds *PageReportRepository) FindPageReportsRedirectingToURL(u string, cid int64, p int) []models.PageReport {
	max := paginationMax
	offset := max * (p - 1)
	uh := Hash(u)
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE redirect_hash = ? AND crawl_id = ? AND crawled = 1
		LIMIT ?,?`

	var pageReports []models.PageReport
	rows, err := ds.DB.Query(query, uh, cid, offset, max)
	if err != nil {
		log.Println(err)
	}

	for rows.Next() {
		p := models.PageReport{}
		err := rows.Scan(&p.Id, &p.URL, &p.Title)
		if err != nil {
			log.Println(err)
			continue
		}

		pageReports = append(pageReports, p)
	}

	return pageReports
}

// GetNumberOfPagesForLinks returns the total number of pages of Links for an specific PageReport.
// this method is used to build a paginator of Links.
func (ds *PageReportRepository) GetNumberOfPagesForLinks(pageReport *models.PageReport, cid int64) int {
	query := `
		SELECT
			count(*)
		FROM links
		WHERE pagereport_id = ? AND crawl_id = ?
	`

	row := ds.DB.QueryRow(query, pageReport.Id, cid)
	var c int
	if err := row.Scan(&c); err != nil {
		log.Printf("GetNumberOfPagesForLinks: %v\n", err)
	}
	var f float64 = float64(c) / float64(paginationMax)
	return int(math.Ceil(f))
}

// GetNumberOfPagesForExternalLinks returns the total number of pages of external Links for an specific PageReport.
// this method is used to build a paginator of external links.
func (ds *PageReportRepository) GetNumberOfPagesForExternalLinks(pageReport *models.PageReport, cid int64) int {
	query := `
		SELECT
			count(*)
		FROM external_links
		WHERE pagereport_id = ? AND crawl_id = ?
	`

	row := ds.DB.QueryRow(query, pageReport.Id, cid)
	var c int
	if err := row.Scan(&c); err != nil {
		log.Printf("GetNumberOfPagesForExternalLinks: %v\n", err)
	}
	var f float64 = float64(c) / float64(paginationMax)
	return int(math.Ceil(f))
}

// GetNumberOfPagesForInlinks returns the total number of pages of in links for an specific PageReport.
// this method is used to build a paginator of in links.
func (ds *PageReportRepository) GetNumberOfPagesForInlinks(pageReport *models.PageReport, cid int64) int {
	h := Hash(pageReport.URL)
	query := `
		SELECT 
			count(pagereports.id)
		FROM links
		LEFT JOIN pagereports ON pagereports.id = links.pagereport_id
		WHERE links.url_hash = ? AND pagereports.crawl_id = ? AND pagereports.crawled = 1
	`

	row := ds.DB.QueryRow(query, h, cid)
	var c int
	if err := row.Scan(&c); err != nil {
		log.Printf("GetNumberOfPagesForInlinks: %v\n", err)
	}
	var f float64 = float64(c) / float64(paginationMax)
	return int(math.Ceil(f))
}

// GetNumberOfPagesForRedirecting returns the total number of pages of redirects for an specific PageReport.
// this method is used to build a paginator of redirects.
func (ds *PageReportRepository) GetNumberOfPagesForRedirecting(pageReport *models.PageReport, cid int64) int {
	h := Hash(pageReport.URL)
	query := `
		SELECT
			count(id)
		FROM pagereports
		WHERE redirect_hash = ? AND crawl_id = ? AND crawled = 1
	`

	row := ds.DB.QueryRow(query, h, cid)
	var c int
	if err := row.Scan(&c); err != nil {
		log.Printf("GetNumberOfPagesForRedirecting: %v\n", err)
	}
	var f float64 = float64(c) / float64(paginationMax)
	return int(math.Ceil(f))

}
