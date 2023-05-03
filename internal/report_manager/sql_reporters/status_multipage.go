package sql_reporters

import (
	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporter_errors"
)

// Creates a MultipageIssueReporter object that contains the SQL query to check for pages
// that redirect to pages that are also redirected somewhere else.
func RedirectChainsReporter(c *models.Crawl) *MultipageIssueReporter {
	query := `
		SELECT
			a.id,
			a.url,
			a.title
		FROM pagereports AS a
		LEFT JOIN pagereports AS b ON a.redirect_hash = b.url_hash
		WHERE a.redirect_hash != "" AND b.redirect_hash  != "" AND a.crawl_id = ? AND b.crawl_id = ?
		AND a.crawled = 1 AND b.crawled = 1`

	return &MultipageIssueReporter{
		Query:      query,
		Parameters: []interface{}{c.Id, c.Id},
		ErrorType:  reporter_errors.ErrorRedirectChain,
	}
}

// Creates a MultipageIssueReporter object that contains the SQL query to check for pages
// that redirect to pages that are redirected back, creating redirection loops.
func RedirectLoopsReporter(c *models.Crawl) *MultipageIssueReporter {
	query := `
		SELECT
			a.id,
			a.url,
			a.title
		FROM pagereports AS a
		INNER JOIN pagereports AS b ON a.redirect_hash = b.url_hash AND b.redirect_hash = a.url_hash
		WHERE a.crawl_id = ? AND b.crawl_id = ?
		AND a.crawled = 1 AND b.crawled = 1`

	return &MultipageIssueReporter{
		Query:      query,
		Parameters: []interface{}{c.Id, c.Id},
		ErrorType:  reporter_errors.ErrorRedirectLoop,
	}
}
