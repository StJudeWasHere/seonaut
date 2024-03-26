package multipage

import (
	"github.com/stjudewashere/seonaut/internal/issues/errors"
	"github.com/stjudewashere/seonaut/internal/models"
)

// Creates a MultipageIssueReporter object that contains the SQL query to check for pages
// that redirect to pages that are also redirected somewhere else.
func (sr *SqlReporter) RedirectChainsReporter(c *models.Crawl) *models.MultipageIssueReporter {
	query := `
		SELECT
			a.id
		FROM pagereports AS a
		LEFT JOIN pagereports AS b ON a.redirect_hash = b.url_hash
		WHERE a.redirect_hash != "" AND b.redirect_hash  != "" AND a.crawl_id = ? AND b.crawl_id = ?
		AND a.crawled = 1 AND b.crawled = 1`

	return &models.MultipageIssueReporter{
		Pstream:   sr.pageReportsQuery(query, c.Id, c.Id),
		ErrorType: errors.ErrorRedirectChain,
	}
}

// Creates a MultipageIssueReporter object that contains the SQL query to check for pages
// that redirect to pages that are redirected back, creating redirection loops.
func (sr *SqlReporter) RedirectLoopsReporter(c *models.Crawl) *models.MultipageIssueReporter {
	query := `
		SELECT
			a.id
		FROM pagereports AS a
		INNER JOIN pagereports AS b ON a.redirect_hash = b.url_hash AND b.redirect_hash = a.url_hash
		WHERE a.crawl_id = ? AND b.crawl_id = ?
		AND a.crawled = 1 AND b.crawled = 1`

	return &models.MultipageIssueReporter{
		Pstream:   sr.pageReportsQuery(query, c.Id, c.Id),
		ErrorType: errors.ErrorRedirectLoop,
	}
}
