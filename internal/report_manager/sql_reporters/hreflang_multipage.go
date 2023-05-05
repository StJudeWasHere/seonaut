package sql_reporters

import (
	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporter_errors"
)

// Creates a MultipageIssueReporter object that contains the SQL query to check for pages that have missing
// hreflang return links.
func (sr *SqlReporter) MissingHrelangReturnLinks(c *models.Crawl) *report_manager.MultipageIssueReporter {
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

	return &report_manager.MultipageIssueReporter{
		Pstream:   sr.pageReportsQuery(query, c.Id),
		ErrorType: reporter_errors.ErrorHreflangsReturnLink,
	}
}

// Creates a MultipageIssueReporter object that contains the SQL query to check for pages that have hreflang
// links to non-canonical pages.
func (sr *SqlReporter) HreflangsToNonCanonical(c *models.Crawl) *report_manager.MultipageIssueReporter {
	query := `
		SELECT
			pagereports.id
		FROM pagereports
		LEFT JOIN hreflangs ON hreflangs.to_hash = pagereports.url_hash AND hreflangs.crawl_id = ?
		WHERE media_type = "text/html" AND status_code >= 200 AND status_code < 300
		AND (canonical IS NOT NULL AND canonical != "" AND canonical != url) AND pagereports.crawl_id = ?
		AND hreflangs.id IS NOT NULL AND crawled = 1`

	return &report_manager.MultipageIssueReporter{
		Pstream:   sr.pageReportsQuery(query, c.Id, c.Id),
		ErrorType: reporter_errors.ErrorHreflangToNonCanonical,
	}
}

// Creates a MultipageIssueReporter object that contains the SQL query to check for pages that have hreflang
// links to non-indexable pages.
func (sr *SqlReporter) HreflangNoindexable(c *models.Crawl) *report_manager.MultipageIssueReporter {
	query := `
		SELECT
			pagereports.id
		FROM pagereports
		WHERE id IN (
			SELECT
				DISTINCT hreflangs.pagereport_id
			FROM hreflangs 
			INNER JOIN pagereports ON hreflangs.pagereport_id = pagereports.id AND hreflangs.crawl_id = pagereports.crawl_id
			WHERE hreflangs.crawl_id = ? and pagereports.noindex = 1 AND pagereports.crawled = 1
		) AND crawled = 1`

	return &report_manager.MultipageIssueReporter{
		Pstream:   sr.pageReportsQuery(query, c.Id),
		ErrorType: reporter_errors.ErrorHreflangNoindexable,
	}
}
