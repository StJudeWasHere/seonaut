package datastore

import (
	"log"

	"github.com/stjudewashere/seonaut/internal/crawler"
	"github.com/stjudewashere/seonaut/internal/helper"
)

func (ds *Datastore) FindPageReportsRedirectingToURL(u string, cid int64) []crawler.PageReport {
	uh := helper.Hash(u)
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE redirect_hash = ? AND crawl_id = ?`

	return ds.pageReportsQuery(query, uh, cid)
}

func (ds *Datastore) FindPageReportsWithEmptyTitle(cid int64) []crawler.PageReport {
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

func (ds *Datastore) Find40xPageReports(cid int64) []crawler.PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE status_code >= 400 AND status_code < 500 AND crawl_id = ?`

	return ds.pageReportsQuery(query, cid)
}

func (ds *Datastore) Find30xPageReports(cid int64) []crawler.PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE status_code >= 300 AND status_code < 400 AND crawl_id = ?`

	return ds.pageReportsQuery(query, cid)
}

func (ds *Datastore) Find50xPageReports(cid int64) []crawler.PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE status_code >= 500 AND crawl_id = ?`

	return ds.pageReportsQuery(query, cid)
}

func (ds *Datastore) FindPageReportsWithLittleContent(cid int64) []crawler.PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE words < 200 AND status_code >= 200 AND status_code < 300 AND media_type = "text/html" AND crawl_id = ?`

	return ds.pageReportsQuery(query, cid)
}

func (ds *Datastore) FindPageReportsWithShortTitle(cid int64) []crawler.PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE length(title) > 0 AND length(title) < 20 AND media_type = "text/html" AND crawl_id = ?`

	return ds.pageReportsQuery(query, cid)
}

func (ds *Datastore) FindPageReportsWithLongTitle(cid int64) []crawler.PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE length(title) > 60 AND media_type = "text/html" AND crawl_id = ?`

	return ds.pageReportsQuery(query, cid)
}

func (ds *Datastore) FindPageReportsWithDuplicatedTitle(cid int64) []crawler.PageReport {
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
		ON d.title = y.title AND d.lang = y.lang
		WHERE media_type = "text/html" AND length(y.title) > 0 AND crawl_id = ?
		AND status_code >= 200 AND status_code < 300 AND (canonical = "" OR canonical = url)`

	return ds.pageReportsQuery(query, cid, cid)
}

func (ds *Datastore) FindPageReportsWithoutH1(cid int64) []crawler.PageReport {
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

func (ds *Datastore) FindPageReportsWithEmptyDescription(cid int64) []crawler.PageReport {
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

func (ds *Datastore) FindPageReportsWithShortDescription(cid int64) []crawler.PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE length(description) > 0 AND length(description) < 80 AND media_type = "text/html" AND crawl_id = ?`

	return ds.pageReportsQuery(query, cid)
}

func (ds *Datastore) FindPageReportsWithLongDescription(cid int64) []crawler.PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE length(description) > 160 AND media_type = "text/html" AND crawl_id = ?`

	return ds.pageReportsQuery(query, cid)
}

func (ds *Datastore) FindPageReportsWithDuplicatedDescription(cid int) []crawler.PageReport {
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
		ON d.description = y.description AND d.lang = y.lang
		WHERE y.media_type = "text/html" AND length(y.description) > 0 AND y.crawl_id = ?
		AND status_code >= 200 AND status_code < 300 AND (canonical = "" OR canonical = url`

	return ds.pageReportsQuery(query, cid, cid)
}

func (ds *Datastore) FindImagesWithNoAlt(cid int64) []crawler.PageReport {
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

func (ds *Datastore) FindPageReportsWithNoLangAttr(cid int64) []crawler.PageReport {
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

func (ds *Datastore) FindPageReportsWithHTTPLinks(cid int64) []crawler.PageReport {
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

func (ds *Datastore) FindMissingHrelangReturnLinks(cid int64) []crawler.PageReport {
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

func (ds *Datastore) FindInLinks(s string, cid int64) []crawler.PageReport {
	hash := helper.Hash(s)
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

func (ds *Datastore) FindPageReportIssues(cid int64, p int, errorType string) []crawler.PageReport {
	max := paginationMax
	offset := max * (p - 1)

	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE id IN (
			SELECT DISTINCT pagereport_id
			FROM issues
			INNER JOIN issue_types ON issue_types.id = issues.issue_type_id
			WHERE issue_types.type = ? and crawl_id  = ?
		) ORDER BY url ASC LIMIT ?, ?`

	return ds.pageReportsQuery(query, errorType, cid, offset, max)
}

func (ds *Datastore) FindRedirectChains(cid int64) []crawler.PageReport {
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

func (ds *Datastore) TooManyLinks(cid int64) []crawler.PageReport {
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

func (ds *Datastore) InternalNoFollowLinks(cid int64) []crawler.PageReport {
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

func (ds *Datastore) FindSitemapPageReports(cid int64) []crawler.PageReport {
	query := `
		SELECT pagereports.id, pagereports.url, pagereports.title
		FROM pagereports
		WHERE media_type = "text/html" AND status_code >= 200 AND status_code < 300
		AND (canonical IS NULL OR canonical = "" OR canonical = url) AND pagereports.crawl_id = ?`

	return ds.pageReportsQuery(query, cid)
}

func (ds *Datastore) FindExternalLinkWitoutNoFollow(cid int64) []crawler.PageReport {
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

func (ds *Datastore) FindCanonicalizedToNonCanonical(cid int64) []crawler.PageReport {
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

func (ds *Datastore) FindRedirectLoops(cid int64) []crawler.PageReport {
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

func (ds *Datastore) FindNotValidHeadingsOrder(cid int64) []crawler.PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE crawl_id = ? AND valid_headings = 0`

	return ds.pageReportsQuery(query, cid)
}

func (ds *Datastore) pageReportsQuery(query string, args ...interface{}) []crawler.PageReport {
	var pageReports []crawler.PageReport
	rows, err := ds.db.Query(query, args...)
	if err != nil {
		log.Println(err)
	}

	for rows.Next() {
		p := crawler.PageReport{}
		err := rows.Scan(&p.Id, &p.URL, &p.Title)
		if err != nil {
			log.Println(err)
			continue
		}

		pageReports = append(pageReports, p)
	}

	return pageReports
}
