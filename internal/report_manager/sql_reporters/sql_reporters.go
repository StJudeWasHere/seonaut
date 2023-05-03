package sql_reporters

import (
	"github.com/stjudewashere/seonaut/internal/models"
)

type MultipageCallback func(c *models.Crawl) *MultipageIssueReporter

type MultipageIssueReporter struct {
	Query      string
	Parameters []interface{}
	ErrorType  int
}

func GetAllReporters() []MultipageCallback {
	return []MultipageCallback{
		// Add status code issue reporters
		RedirectChainsReporter,
		RedirectLoopsReporter,

		// Add title issue reporters
		DuplicatedTitleReporter,

		// Add description issue reporters
		DuplicatedDescriptionReporter,

		// Add link issue reporters
		OrphanPagesReporter,
		NoFollowIndexableReporter,
		FollowNoFollowReporter,

		// Add hreflang reporters
		MissingHrelangReturnLinks,
		HreflangsToNonCanonical,
		HreflangNoindexable,

		// Add canonical issue reporters
		CanonicalizedToNonCanonical,
	}
}
