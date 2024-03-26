package multipage

import (
	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/models"
)

// Creates a MultipageIssueReporter object that contains the SQL query to check for pages that have missing
// hreflang return links.
func (sr *SqlReporter) MissingHrelangReturnLinks(c *models.Crawl) *models.MultipageIssueReporter {
	query := `
		SELECT
			distinct pagereports.id
		FROM hreflangs
		LEFT JOIN hreflangs b ON hreflangs.crawl_id = b.crawl_id and hreflangs.from_hash = b.to_hash
		LEFT JOIN pagereports ON hreflangs.pagereport_id = pagereports.id
		WHERE  hreflangs.crawl_id = ?
			AND hreflangs.to_lang != "x-default"
			AND pagereports.status_code < 300
			AND b.id IS NULL
			AND (pagereports.canonical = "" OR pagereports.canonical = pagereports.URL)
			AND pagereports.crawled = 1`

	return &models.MultipageIssueReporter{
		Pstream:   sr.pageReportsQuery(query, c.Id),
		ErrorType: errors.ErrorHreflangsReturnLink,
	}
}

// Creates a MultipageIssueReporter object that contains the SQL query to check for pages that have hreflang
// links to non-canonical pages.
func (sr *SqlReporter) HreflangsToNonCanonical(c *models.Crawl) *models.MultipageIssueReporter {
	query := `
		SELECT
			pagereports.id
		FROM pagereports
		LEFT JOIN hreflangs ON hreflangs.to_hash = pagereports.url_hash AND hreflangs.crawl_id = ?
		WHERE media_type = "text/html"
			AND status_code >= 200
			AND status_code < 300
			AND (canonical IS NOT NULL AND canonical != "" AND canonical != url)
			AND pagereports.crawl_id = ?
			AND hreflangs.id IS NOT NULL
			AND crawled = 1`

	return &models.MultipageIssueReporter{
		Pstream:   sr.pageReportsQuery(query, c.Id, c.Id),
		ErrorType: errors.ErrorHreflangToNonCanonical,
	}
}

// Creates a MultipageIssueReporter object that contains the SQL query to check for pages that have hreflang
// links to non-indexable pages.
func (sr *SqlReporter) HreflangNoindexable(c *models.Crawl) *models.MultipageIssueReporter {
	query := `
		SELECT
			pagereports.id
		FROM pagereports
		WHERE id IN (
			SELECT
				DISTINCT hreflangs.pagereport_id
			FROM hreflangs 
			INNER JOIN pagereports ON hreflangs.pagereport_id = pagereports.id
				AND hreflangs.crawl_id = pagereports.crawl_id
			WHERE hreflangs.crawl_id = ?
				AND pagereports.noindex = 1
				AND pagereports.crawled = 1
		) AND crawled = 1`

	return &models.MultipageIssueReporter{
		Pstream:   sr.pageReportsQuery(query, c.Id),
		ErrorType: errors.ErrorHreflangNoindexable,
	}
}

// Creates a MultipageIssueReporter object that contains the SQL query to check for pages
// that have hreflang links pointing to redirects.
func (sr *SqlReporter) HreflangToRedirect(c *models.Crawl) *models.MultipageIssueReporter {
	query := `
		SELECT
			id
		FROM pagereports AS pr
		INNER JOIN hreflangs AS hl ON pr.id = hl.pagereport_id
		INNER JOIN pagereports AS pr2 ON hl.to_hash = pr2.url_hash
		WHERE pr.crawl_id = ?
			AND pr2.status_code >= 300
			AND pr2.status_code < 400;`

	return &models.MultipageIssueReporter{
		Pstream:   sr.pageReportsQuery(query, c.Id),
		ErrorType: errors.ErrorHreflangToRedirect,
	}
}

// Creates a MultipageIssueReporter object that contains the SQL query to check for pages
// that have hreflang links pointing to error pages with status code 40x or 50x.
func (sr *SqlReporter) HreflangToError(c *models.Crawl) *models.MultipageIssueReporter {
	query := `
		SELECT
			id
		FROM pagereports AS pr
		INNER JOIN hreflangs AS hl ON pr.id = hl.pagereport_id
		INNER JOIN pagereports AS pr2 ON hl.to_hash = pr2.url_hash
		WHERE pr.crawl_id = ?
			AND pr2.status_code >= 400;`

	return &models.MultipageIssueReporter{
		Pstream:   sr.pageReportsQuery(query, c.Id),
		ErrorType: errors.ErrorHreflangToError,
	}
}

// Creates a MultipageIssueReporter object that contains the SQL query to check for pages
// that are referenced from hreflang with more than one language.
func (sr *SqlReporter) MultipleLangReference(c *models.Crawl) *models.MultipageIssueReporter {
	query := `
		SELECT pagereports.id
		FROM hreflangs 
		LEFT JOIN pagereports
			ON to_hash = pagereports.url_hash AND hreflangs.crawl_id = ?
		WHERE pagereports.crawl_id = ? AND hreflangs.to_lang != "x-default" 
		GROUP BY to_hash, pagereports.id
		HAVING count(distinct to_lang) > 1`

	return &models.MultipageIssueReporter{
		Pstream:   sr.pageReportsQuery(query, c.Id, c.Id),
		ErrorType: errors.ErrorMultipleLangReference,
	}
}
