package multipage

import (
	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/models"
)

// Creates a MultipageIssueReporter object that contains the SQL query to check for pages
// that have the same description, taking into account the status code and language.
func (sr *SqlReporter) DuplicatedDescriptionReporter(c *models.Crawl) *models.MultipageIssueReporter {
	query := `
		SELECT
			y.id
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

	return &models.MultipageIssueReporter{
		Pstream:   sr.pageReportsQuery(query, c.Id, c.Id),
		ErrorType: errors.ErrorDuplicatedDescription,
	}
}
