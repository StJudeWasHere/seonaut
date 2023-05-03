package sql_reporters

import (
	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporter_errors"
)

// Creates a MultipageIssueReporter object that contains the SQL query to check for orphan pages.
// Pages with no incoming links are considered orphan pages.
func (sr *SqlReporter) OrphanPagesReporter(c *models.Crawl) *report_manager.MultipageIssueReporter {
	query := `
		SELECT
			pagereports.id,
			pagereports.url,
			pagereports.title
		FROM pagereports
		LEFT JOIN links ON pagereports.url_hash = links.url_hash and pagereports.crawl_id = links.crawl_id
		WHERE pagereports.media_type = "text/html" AND links.url IS NULL AND pagereports.crawl_id = ?`

	return &report_manager.MultipageIssueReporter{
		Pstream:   sr.pageReportsQuery(query, c.Id),
		ErrorType: reporter_errors.ErrorOrphan,
	}
}
