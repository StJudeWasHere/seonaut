package sql_reporters

import (
	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporter_errors"
)

// Creates a MultipageIssueReporter object that contains the SQL query to check for pages
// that are canonicalized to non-canonical pages.
func CanonicalizedToNonCanonical(c *models.Crawl) *MultipageIssueReporter {
	query := `
		SELECT
			a.id,
			a.url,
			a.title
		FROM pagereports AS a
		INNER JOIN pagereports AS b ON a.url = b.canonical
		WHERE a.crawl_id = ? AND b.crawl_id = ? AND a.canonical != "" AND a.canonical != a.url
		AND a.crawled = 1 AND b.crawled = 1`

	return &MultipageIssueReporter{
		Query:      query,
		Parameters: []interface{}{c.Id, c.Id},
		ErrorType:  reporter_errors.ErrorCanonicalizedToNonCanonical,
	}
}
