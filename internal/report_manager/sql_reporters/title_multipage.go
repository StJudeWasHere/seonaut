package sql_reporters

import (
	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporter_errors"
)

// Creates a MultipageIssueReporter object that contains the SQL query to check for pages with identical titles.
// It considers factors such as the HTTP status code, media type and whether they are canonical or not.
func DuplicatedTitleReporter(c *models.Crawl) *MultipageIssueReporter {
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

	return &MultipageIssueReporter{
		Query:      query,
		Parameters: []interface{}{c.Id, c.Id},
		ErrorType:  reporter_errors.ErrorDuplicatedTitle,
	}
}
