package datastore

import (
	"log"

	"github.com/stjudewashere/seonaut/internal/pagereport"
)

func (ds *Datastore) FindPageReportsWithEmptyTitle(cid int64) <-chan *pagereport.PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE (title = "" OR title IS NULL) AND media_type = "text/html"
		AND status_code >=200 AND status_code < 300 AND crawl_id = ? AND crawled = 1`

	return ds.pageReportsQuery(query, cid)
}

func (ds *Datastore) Find40xPageReports(cid int64) <-chan *pagereport.PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE status_code >= 400 AND status_code < 500 AND crawl_id = ? AND crawled = 1`

	return ds.pageReportsQuery(query, cid)
}

func (ds *Datastore) Find30xPageReports(cid int64) <-chan *pagereport.PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE status_code >= 300 AND status_code < 400 AND crawl_id = ? AND crawled = 1`

	return ds.pageReportsQuery(query, cid)
}

func (ds *Datastore) Find50xPageReports(cid int64) <-chan *pagereport.PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE status_code >= 500 AND crawl_id = ? AND crawled = 1`

	return ds.pageReportsQuery(query, cid)
}

func (ds *Datastore) FindPageReportsWithLittleContent(cid int64) <-chan *pagereport.PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE words < 200 AND status_code >= 200 AND status_code < 300 AND media_type = "text/html" AND crawl_id = ? AND crawled = 1`

	return ds.pageReportsQuery(query, cid)
}

func (ds *Datastore) FindPageReportsWithShortTitle(cid int64) <-chan *pagereport.PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE length(title) > 0 AND length(title) < 20 AND media_type = "text/html" AND crawl_id = ? AND crawled = 1`

	return ds.pageReportsQuery(query, cid)
}

func (ds *Datastore) FindPageReportsWithLongTitle(cid int64) <-chan *pagereport.PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE length(title) > 60 AND media_type = "text/html" AND crawl_id = ? AND crawled = 1`

	return ds.pageReportsQuery(query, cid)
}

func (ds *Datastore) FindPageReportsWithDuplicatedTitle(cid int64) <-chan *pagereport.PageReport {
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
			AND status_code < 300 AND (canonical = "" OR canonical = url) AND crawled = 1
			GROUP BY title, lang
			HAVING c > 1
		) d 
		ON d.title = y.title AND d.lang = y.lang
		WHERE media_type = "text/html" AND length(y.title) > 0 AND crawl_id = ?
		AND status_code >= 200 AND status_code < 300 AND (canonical = "" OR canonical = url) AND crawled = 1`

	return ds.pageReportsQuery(query, cid, cid)
}

func (ds *Datastore) FindPageReportsWithoutH1(cid int64) <-chan *pagereport.PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE (h1 = "" OR h1 IS NULL) AND media_type = "text/html"
		AND status_code >= 200 AND status_code < 300 AND crawl_id = ? AND crawled = 1`

	return ds.pageReportsQuery(query, cid)
}

func (ds *Datastore) FindPageReportsWithEmptyDescription(cid int64) <-chan *pagereport.PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE (description = "" OR description IS NULL) AND media_type = "text/html"
		AND status_code >= 200 AND status_code < 300 AND crawl_id = ? AND crawled = 1`

	return ds.pageReportsQuery(query, cid)
}

func (ds *Datastore) FindPageReportsWithShortDescription(cid int64) <-chan *pagereport.PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE length(description) > 0 AND length(description) < 80 AND media_type = "text/html" AND crawl_id = ? AND crawled = 1`

	return ds.pageReportsQuery(query, cid)
}

func (ds *Datastore) FindPageReportsWithLongDescription(cid int64) <-chan *pagereport.PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE length(description) > 160 AND media_type = "text/html" AND crawl_id = ? AND crawled = 1`

	return ds.pageReportsQuery(query, cid)
}

func (ds *Datastore) FindPageReportsWithDuplicatedDescription(cid int64) <-chan *pagereport.PageReport {
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
			AND status_code < 300 AND (canonical = "" OR canonical = url) AND crawled = 1
			GROUP BY description, lang
			HAVING c > 1
		) d 
		ON d.description = y.description AND d.lang = y.lang
		WHERE y.media_type = "text/html" AND length(y.description) > 0 AND y.crawl_id = ?
		AND status_code >= 200 AND status_code < 300 AND (canonical = "" OR canonical = url AND crawled = 1)`

	return ds.pageReportsQuery(query, cid, cid)
}

func (ds *Datastore) FindImagesWithNoAlt(cid int64) <-chan *pagereport.PageReport {
	query := `
		SELECT
			pagereports.id,
			pagereports.url,
			pagereports.title
		FROM pagereports
		LEFT JOIN images ON images.pagereport_id = pagereports.id
		WHERE images.alt = "" AND pagereports.crawl_id = ? AND pagereports.crawled = 1
		GROUP BY pagereports.id`

	return ds.pageReportsQuery(query, cid)
}

func (ds *Datastore) FindPageReportsWithNoLangAttr(cid int64) <-chan *pagereport.PageReport {
	query := `
		SELECT
			pagereports.id,
			pagereports.url,
			pagereports.title
		FROM pagereports
		WHERE (pagereports.lang = "" OR pagereports.lang = null) and media_type = "text/html"
		AND pagereports.crawl_id = ? AND crawled = 1`

	return ds.pageReportsQuery(query, cid)
}

func (ds *Datastore) FindPageReportsWithHTTPLinks(cid int64) <-chan *pagereport.PageReport {
	query := `
		SELECT
			pagereports.id,
			pagereports.url,
			pagereports.title
		FROM pagereports
		LEFT JOIN links ON links.pagereport_id = pagereports.id
		WHERE pagereports.scheme = "https" AND links.scheme = "http" AND crawled = 1
		AND pagereports.crawl_id = ?
		GROUP BY links.pagereport_id
		HAVING count(links.pagereport_id) > 1`

	return ds.pageReportsQuery(query, cid)
}

func (ds *Datastore) FindMissingHrelangReturnLinks(cid int64) <-chan *pagereport.PageReport {
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
		AND (pagereports.canonical = "" OR pagereports.canonical = pagereports.URL) AND pagereports.crawled = 1`

	return ds.pageReportsQuery(query, cid)
}

func (ds *Datastore) FindRedirectChains(cid int64) <-chan *pagereport.PageReport {
	query := `
		SELECT
			a.id,
			a.url,
			a.title
		FROM pagereports AS a
		LEFT JOIN pagereports AS b ON a.redirect_hash = b.url_hash
		WHERE a.redirect_hash != "" AND b.redirect_hash  != "" AND a.crawl_id = ? AND b.crawl_id = ?
		AND a.crawled = 1 AND b.crawled = 1`

	return ds.pageReportsQuery(query, cid, cid)
}

func (ds *Datastore) TooManyLinks(cid int64) <-chan *pagereport.PageReport {
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
		WHERE pagereports.crawl_id = ? and l > 100 AND crawled = 1
	`

	return ds.pageReportsQuery(query, cid, cid)
}

func (ds *Datastore) InternalNoFollowLinks(cid int64) <-chan *pagereport.PageReport {
	query := `
		SELECT pagereports.id, pagereports.url, pagereports.title
		FROM pagereports 
		INNER JOIN (
			SELECT
				DISTINCT links.pagereport_id
			FROM links
			WHERE links.nofollow = 1 AND links.crawl_id = ?
		) AS b ON b.pagereport_id = pagereports.id
		WHERE pagereports.crawl_id = ? AND crawled = 1`

	return ds.pageReportsQuery(query, cid, cid)
}

func (ds *Datastore) InternalNoFollowIndexableLinks(cid int64) <-chan *pagereport.PageReport {
	query := `
		SELECT
			pagereports.id,
			pagereports.url,
			pagereports.title
		FROM pagereports
		INNER JOIN (
			SELECT
				a.pagereport_id
			FROM (
				SELECT
					DISTINCT links.pagereport_id,
					links.url_hash
				FROM links
				WHERE links.crawl_id = ? AND links.nofollow = 1
			) AS a INNER JOIN pagereports ON a.url_hash = pagereports.url_hash
			WHERE pagereports.noindex = 0 AND pagereports.crawled = 1 AND pagereports.crawl_id = ?
		) AS b
		ON b.pagereport_id = pagereports.id
	`

	return ds.pageReportsQuery(query, cid, cid)
}

func (ds *Datastore) NoIndexable(cid int64) <-chan *pagereport.PageReport {
	query := `
		SELECT pagereports.id, pagereports.url, pagereports.title
		FROM pagereports 
		WHERE pagereports.noindex = 1 AND pagereports.crawl_id = ?`

	return ds.pageReportsQuery(query, cid)
}

func (ds *Datastore) HreflangNoindexable(cid int64) <-chan *pagereport.PageReport {
	query := `
		SELECT pagereports.id, pagereports.url, pagereports.title
		FROM pagereports
		WHERE id IN (
			SELECT DISTINCT hreflangs.pagereport_id
			FROM hreflangs 
			INNER JOIN pagereports ON hreflangs.pagereport_id = pagereports.id AND hreflangs.crawl_id = pagereports.crawl_id
			WHERE hreflangs.crawl_id = ? and pagereports.noindex = 1 AND pagereports.crawled = 1
		) AND crawled = 1`

	return ds.pageReportsQuery(query, cid)
}

func (ds *Datastore) FindExternalLinkWitoutNoFollow(cid int64) <-chan *pagereport.PageReport {
	query := `
		SELECT
			pagereports.id,
			pagereports.url,
			pagereports.title
		FROM pagereports
		INNER JOIN external_links ON pagereports.id = external_links.pagereport_id
		WHERE external_links.nofollow = 0 AND pagereports.crawl_id = ? AND pagereports.crawled = 1
		GROUP BY pagereports.id`

	return ds.pageReportsQuery(query, cid)
}

func (ds *Datastore) FindCanonicalizedToNonCanonical(cid int64) <-chan *pagereport.PageReport {
	query := `
		SELECT
			a.id,
			a.url,
			a.title
		FROM pagereports AS a
		INNER JOIN pagereports AS b ON a.url = b.canonical
		WHERE a.crawl_id = ? AND b.crawl_id = ? AND a.canonical != "" AND a.canonical != a.url
		AND a.crawled = 1 AND b.crawled = 1`

	return ds.pageReportsQuery(query, cid, cid)
}

func (ds *Datastore) FindRedirectLoops(cid int64) <-chan *pagereport.PageReport {
	query := `
		SELECT
			a.id,
			a.url,
			a.title
		FROM pagereports AS a
		INNER JOIN pagereports AS b ON a.redirect_hash = b.url_hash AND b.redirect_hash = a.url_hash
		WHERE a.crawl_id = ? AND b.crawl_id = ?
		AND a.crawled = 1 AND b.crawled = 1`

	return ds.pageReportsQuery(query, cid, cid)
}

func (ds *Datastore) FindNotValidHeadingsOrder(cid int64) <-chan *pagereport.PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE crawl_id = ? AND valid_headings = 0 AND crawled = 1`

	return ds.pageReportsQuery(query, cid)
}

func (ds *Datastore) FindHreflangsToNonCanonical(cid int64) <-chan *pagereport.PageReport {
	query := `
		SELECT
			pagereports.id,
			pagereports.url,
			pagereports.title
		FROM pagereports
		LEFT JOIN hreflangs ON hreflangs.to_hash = pagereports.url_hash AND hreflangs.crawl_id = ?
		WHERE media_type = "text/html" AND status_code >= 200 AND status_code < 300
		AND (canonical IS NOT NULL AND canonical != "" AND canonical != url) AND pagereports.crawl_id = ?
		AND hreflangs.id IS NOT NULL AND crawled = 1`

	return ds.pageReportsQuery(query, cid, cid)
}

func (ds *Datastore) FindBlockedByRobotstxt(cid int64) <-chan *pagereport.PageReport {
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE crawl_id = ? AND robotstxt_blocked = 1 AND crawled = 0`

	return ds.pageReportsQuery(query, cid)
}

func (ds *Datastore) FindOrphanPages(cid int64) <-chan *pagereport.PageReport {
	query := `
		SELECT
			pagereports.id,
			pagereports.url,
			pagereports.title
		FROM pagereports
		LEFT JOIN links ON pagereports.url_hash = links.url_hash and pagereports.crawl_id = links.crawl_id
		WHERE pagereports.media_type = "text/html" AND links.url IS NULL AND pagereports.crawl_id = ?`

	return ds.pageReportsQuery(query, cid)
}

// Finds non-indexable pagereports that are included in the sitemap.
func (ds *Datastore) FindNoIndexInSitemap(cid int64) <-chan *pagereport.PageReport {
	query := `
		SELECT
			pagereports.id,
			pagereports.url,
			pagereports.title
		FROM pagereports
		WHERE pagereports.crawl_id = ? AND pagereports.noindex = 1 AND pagereports.in_sitemap = 1`

	return ds.pageReportsQuery(query, cid)
}

// Finds pagereports that are blocked by robots.txt and are included in the sitemap.
func (ds *Datastore) FindBlockedInSitemap(cid int64) <-chan *pagereport.PageReport {
	query := `
		SELECT
			pagereports.id,
			pagereports.url,
			pagereports.title
		FROM pagereports
		WHERE pagereports.crawl_id = ? AND pagereports.robotstxt_blocked = 1 AND pagereports.in_sitemap = 1`

	return ds.pageReportsQuery(query, cid)
}

// Finds non-canonical pagereports that are included in the sitemap.
func (ds *Datastore) FindNonCanonicalInSitemap(cid int64) <-chan *pagereport.PageReport {
	query := `
		SELECT
			pagereports.id,
			pagereports.url,
			pagereports.title
		FROM pagereports
		WHERE pagereports.crawl_id = ? AND pagereports.canonical != "" AND pagereports.canonical != pagereports.url`

	return ds.pageReportsQuery(query, cid)
}

// Finds pages with index and nonindex incoming links.
func (ds *Datastore) FindIncomingIndexNoIndex(cid int64) <-chan *pagereport.PageReport {
	query := `
		SELECT
			pagereports.id,
			pagereports.url,
			pagereports.title
		FROM pagereports WHERE crawl_id = ? and url_hash in (
			SELECT
				url_hash
			FROM links
			WHERE crawl_id = ?
			GROUP BY url_hash
			HAVING COUNT(DISTINCT nofollow) > 1
		)
	`

	return ds.pageReportsQuery(query, cid, cid)
}

// Finds pages with invalid lang attribute.
func (ds *Datastore) FindInvalidLang(cid int64) <-chan *pagereport.PageReport {
	query := `
		SELECT
			pagereports.id,
			pagereports.url,
			pagereports.title
		FROM pagereports WHERE crawl_id = ? AND lang <> ""
			AND valid_lang = 0 AND media_type = "text/html" AND crawled = 1
	`

	return ds.pageReportsQuery(query, cid)
}

func (ds *Datastore) pageReportsQuery(query string, args ...interface{}) <-chan *pagereport.PageReport {
	prStream := make(chan *pagereport.PageReport)

	go func() {
		defer close(prStream)

		rows, err := ds.db.Query(query, args...)
		if err != nil {
			log.Println(err)
		}

		for rows.Next() {
			p := &pagereport.PageReport{}
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
