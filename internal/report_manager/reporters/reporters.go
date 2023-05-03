package reporters

import (
	"github.com/stjudewashere/seonaut/internal/models"
)

type PageIssueReporter struct {
	Callback  func(*models.PageReport) bool
	ErrorType int
}

// Returns an slice with all available PageIssueReporters.
func GetAllReporters() []*PageIssueReporter {
	return []*PageIssueReporter{
		// Add status code issue reporters
		NewStatus30xReporter(),
		NewStatus40xReporter(),
		NewStatus50xReporter(),

		// Add title issue reporters
		NewEmptyTitleReporter(),
		NewShortTitleReporter(),
		NewShortTitleReporter(),

		// Add description issue reporters
		NewEmptyDescriptionReporter(),
		NewShortDescriptionReporter(),
		NewLongDescriptionReporter(),

		// Add indexability issue reporters
		NewNoIndexableReporter(),
		NewBlockedByRobotstxtReporter(),
		NewNoIndexInSitemapReporter(),
		NewSitemapAndBlockedReporter(),
		NewNonCanonicalInSitemapReporter(),

		// Add link issue reporters
		NewTooManyLinksReporter(),
		NewInternalNoFollowLinksReporter(),
		NewExternalLinkWitoutNoFollowReporter(),
		NewHTTPLinksReporter(),

		// Add image issue reporters
		NewAltTextReporter(),

		// Add language issue reporters
		NewInvalidLangReporter(),
		NewMissingLangReporter(),

		// Add heading issue reporters
		NewNoH1Reporter(),
		NewValidHeadingsOrderReporter(),

		// Add content issue reporters
		NewLittleContentReporter(),
	}
}
