package sql_reporters

import (
	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporter_errors"
)

// Creates a MultipageIssueReporter object that contains the SQL query to check for pages with identical titles.
// It considers factors such as the HTTP status code, media type and whether they are canonical or not.
func (sr *SqlReporter) DuplicatedContent(c *models.Crawl) *report_manager.MultipageIssueReporter {
	query := `
		SELECT id
		FROM pagereports
		WHERE body_hash IN (
			SELECT body_hash
			FROM pagereports
			WHERE crawl_id = ? AND media_type = "text/html" AND body_hash <> ""
			GROUP BY body_hash
			HAVING COUNT(*) > 1
		) AND crawl_id = ? AND media_type = "text/html" AND body_hash <> ""`

	return &report_manager.MultipageIssueReporter{
		Pstream:   sr.pageReportsQuery(query, c.Id, c.Id),
		ErrorType: reporter_errors.ErrorDuplicatedContent,
	}
}
