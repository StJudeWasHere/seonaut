package reporters

import (
	"github.com/stjudewashere/seonaut/internal/report_manager"
)

// Returns an slice with all available report_manager.PageIssueReporters.
func GetAllReporters() []*report_manager.PageIssueReporter {
	return []*report_manager.PageIssueReporter{
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
		NewCanonicalMultipleTagsReporter(),
		NewCanonicalRelativeURLReporter(),
		NewCanonicalMismatch(),

		// Add link issue reporters
		NewTooManyLinksReporter(),
		NewInternalNoFollowLinksReporter(),
		NewExternalLinkWitoutNoFollowReporter(),
		NewHTTPLinksReporter(),
		NewDeadendReporter(),

		// Add image issue reporters
		NewAltTextReporter(),

		// Add language issue reporters
		NewInvalidLangReporter(),
		NewMissingLangReporter(),
		NewHreflangXDefaultMissing(),
		NewHreflangMissingSelfReference(),
		NewHreflangMismatchingLang(),
		NewHreflangRelativeURL(),

		// Add heading issue reporters
		NewNoH1Reporter(),
		NewValidHeadingsOrderReporter(),

		// Add content issue reporters
		NewLittleContentReporter(),

		// Add scheme issue reporters
		NewHTTPSchemeReporter(),

		// Add security issue reporters
		NewMissingHSTSHeaderReporter(),
		NewMissingCSPReporter(),
	}
}
