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
		WHERE a.crawl_id = ? AND b.crawl_id = ? AND a.canonical != "" AND a.canonical != a.url
		AND a.crawled = 1 AND b.crawled = 1`

	return &report_manager.MultipageIssueReporter{
		Pstream:   sr.pageReportsQuery(query, c.Id, c.Id),
		ErrorType: reporter_errors.ErrorCanonicalizedToNonCanonical,
	}
}
