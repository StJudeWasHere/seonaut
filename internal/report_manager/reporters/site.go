package reporters

import (
	"github.com/stjudewashere/seonaut/internal/models"
)

type DatabaseReporter interface {
	PageReportsQuery(query string, args ...interface{}) <-chan *models.PageReport
}

func FindPageReportsWithDuplicatedTitle(ds DatabaseReporter, c *models.Crawl) <-chan *models.PageReport {
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

	return ds.PageReportsQuery(query, c.Id, c.Id)
}

func FindPageReportsWithDuplicatedDescription(ds DatabaseReporter, c *models.Crawl) <-chan *models.PageReport {
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

	return ds.PageReportsQuery(query, c.Id, c.Id)
}

func FindMissingHrelangReturnLinks(ds DatabaseReporter, c *models.Crawl) <-chan *models.PageReport {
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

	return ds.PageReportsQuery(query, c.Id)
}

func FindRedirectChains(ds DatabaseReporter, c *models.Crawl) <-chan *models.PageReport {
	query := `
		SELECT
			a.id,
			a.url,
			a.title
		FROM pagereports AS a
		LEFT JOIN pagereports AS b ON a.redirect_hash = b.url_hash
		WHERE a.redirect_hash != "" AND b.redirect_hash  != "" AND a.crawl_id = ? AND b.crawl_id = ?
		AND a.crawled = 1 AND b.crawled = 1`

	return ds.PageReportsQuery(query, c.Id, c.Id)
}

func InternalNoFollowIndexableLinks(ds DatabaseReporter, c *models.Crawl) <-chan *models.PageReport {
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

	return ds.PageReportsQuery(query, c.Id, c.Id)
}

func FindHreflangNoindexable(ds DatabaseReporter, c *models.Crawl) <-chan *models.PageReport {
	query := `
		SELECT pagereports.id, pagereports.url, pagereports.title
		FROM pagereports
		WHERE id IN (
			SELECT DISTINCT hreflangs.pagereport_id
			FROM hreflangs 
			INNER JOIN pagereports ON hreflangs.pagereport_id = pagereports.id AND hreflangs.crawl_id = pagereports.crawl_id
			WHERE hreflangs.crawl_id = ? and pagereports.noindex = 1 AND pagereports.crawled = 1
		) AND crawled = 1`

	return ds.PageReportsQuery(query, c.Id)
}

func FindCanonicalizedToNonCanonical(ds DatabaseReporter, c *models.Crawl) <-chan *models.PageReport {
	query := `
		SELECT
			a.id,
			a.url,
			a.title
		FROM pagereports AS a
		INNER JOIN pagereports AS b ON a.url = b.canonical
		WHERE a.crawl_id = ? AND b.crawl_id = ? AND a.canonical != "" AND a.canonical != a.url
		AND a.crawled = 1 AND b.crawled = 1`

	return ds.PageReportsQuery(query, c.Id, c.Id)
}

func FindRedirectLoops(ds DatabaseReporter, c *models.Crawl) <-chan *models.PageReport {
	query := `
		SELECT
			a.id,
			a.url,
			a.title
		FROM pagereports AS a
		INNER JOIN pagereports AS b ON a.redirect_hash = b.url_hash AND b.redirect_hash = a.url_hash
		WHERE a.crawl_id = ? AND b.crawl_id = ?
		AND a.crawled = 1 AND b.crawled = 1`

	return ds.PageReportsQuery(query, c.Id, c.Id)
}

func FindHreflangsToNonCanonical(ds DatabaseReporter, c *models.Crawl) <-chan *models.PageReport {
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

	return ds.PageReportsQuery(query, c.Id, c.Id)
}

func FindOrphanPages(ds DatabaseReporter, c *models.Crawl) <-chan *models.PageReport {
	query := `
		SELECT
			pagereports.id,
			pagereports.url,
			pagereports.title
		FROM pagereports
		LEFT JOIN links ON pagereports.url_hash = links.url_hash and pagereports.crawl_id = links.crawl_id
		WHERE pagereports.media_type = "text/html" AND links.url IS NULL AND pagereports.crawl_id = ?`

	return ds.PageReportsQuery(query, c.Id)
}

// Finds pages with index and nonindex incoming links.
func FindIncomingIndexNoIndex(ds DatabaseReporter, c *models.Crawl) <-chan *models.PageReport {
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

	return ds.PageReportsQuery(query, c.Id, c.Id)
}
