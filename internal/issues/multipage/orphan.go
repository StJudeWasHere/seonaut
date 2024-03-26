package multipage

import (
	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/models"
)

// Creates a MultipageIssueReporter object that contains the SQL query to check for orphan pages.
// Pages with no incoming links are considered orphan pages.
func (sr *SqlReporter) OrphanPagesReporter(c *models.Crawl) *models.MultipageIssueReporter {
	query := `
		SELECT
			pagereports.id
		FROM pagereports
		LEFT JOIN links ON pagereports.url_hash = links.url_hash and pagereports.crawl_id = links.crawl_id
		WHERE pagereports.media_type = "text/html" AND links.url IS NULL AND pagereports.crawl_id = ?`

	return &models.MultipageIssueReporter{
		Pstream:   sr.pageReportsQuery(query, c.Id),
		ErrorType: errors.ErrorOrphan,
	}
}
