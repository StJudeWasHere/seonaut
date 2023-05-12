package sql_reporters

import (
	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporter_errors"
)

// Creates a MultipageIssueReporter object that contains the SQL query to check for pages
// that are canonicalized to non-canonical pages.
func (sr *SqlReporter) CanonicalizedToNonCanonical(c *models.Crawl) *report_manager.MultipageIssueReporter {
	query := `
		SELECT
			a.id
		FROM pagereports AS a
		INNER JOIN pagereports AS b ON a.url = b.canonical
		WHERE a.crawl_id = ?
			AND b.crawl_id = ?
			AND a.canonical != ""
			AND a.canonical != a.url
			AND a.crawled = 1
			AND b.crawled = 1`

	return &report_manager.MultipageIssueReporter{
		Pstream:   sr.pageReportsQuery(query, c.Id, c.Id),
		ErrorType: reporter_errors.ErrorCanonicalizedToNonCanonical,
	}
}

// Creates a MultipageIssueReporter object that contains the SQL query to check for pages
// that are canonicalized to non-indexable pages.
func (sr *SqlReporter) CanonicalizedToNonIndexable(c *models.Crawl) *report_manager.MultipageIssueReporter {
	query := `
		SELECT id
		FROM pagereports
		WHERE
			crawl_id = ? AND noindex = 1
		AND canonical IN (
			SELECT url
			FROM pagereports
			WHERE 
			crawl_id = ? AND canonical != "" AND canonical != url
		)`

	return &report_manager.MultipageIssueReporter{
		Pstream:   sr.pageReportsQuery(query, c.Id, c.Id),
		ErrorType: reporter_errors.ErrorCanonicalizedToNonIndexable,
	}
}

// Creates a MultipageIssueReporter object that contains the SQL query to check for pages
// that are canonicalized to redirects.
func (sr *SqlReporter) CanonicalizedToRedirect(c *models.Crawl) *report_manager.MultipageIssueReporter {
	query := `
		SELECT
			pr.id
		FROM pagereports AS pr
		INNER JOIN pagereports AS pr2 ON pr.canonical = pr2.url
		WHERE pr.crawl_id = ?
			AND pr2.crawl_id = ?
			AND pr.canonical != pr.url
			AND pr2.status_code >= 300
			AND pr2.status_code < 400;`

	return &report_manager.MultipageIssueReporter{
		Pstream:   sr.pageReportsQuery(query, c.Id),
		ErrorType: reporter_errors.ErrorCanonicalizedToRedirect,
	}
}

// Creates a MultipageIssueReporter object that contains the SQL query to check for pages
// that are canonicalized to error pages with status code in the 40x and 50x range.
func (sr *SqlReporter) CanonicalizedToError(c *models.Crawl) *report_manager.MultipageIssueReporter {
	query := `
		SELECT
			pr.id
		FROM pagereports AS pr
		INNER JOIN pagereports AS pr2 ON pr.canonical = pr2.url
		WHERE pr.crawl_id = ?
			AND pr2.crawl_id = ?
			AND pr.canonical != pr.url
			AND pr2.status_code >= 400;`

	return &report_manager.MultipageIssueReporter{
		Pstream:   sr.pageReportsQuery(query, c.Id),
		ErrorType: reporter_errors.ErrorCanonicalizedToError,
	}
}
