package page

import "github.com/stjudewashere/seonaut/internal/models"

// Returns an slice with all available report_manager.PageIssueReporters.
func GetAllReporters() []*models.PageIssueReporter {
	return []*models.PageIssueReporter{
		// Add status code issue reporters
		NewStatus30xReporter(),
		NewStatus40xReporter(),
		NewStatus50xReporter(),

		// Add title issue reporters
		NewEmptyTitleReporter(),
		NewShortTitleReporter(),
		NewLongTitleReporter(),
		NewMultipleTitleTagsReporter(),

		// Add description issue reporters
		NewEmptyDescriptionReporter(),
		NewShortDescriptionReporter(),
		NewLongDescriptionReporter(),
		NewMultipleDescriptionTagsReporter(),

		// Add indexability issue reporters
		NewNoIndexableReporter(),
		NewBlockedByRobotstxtReporter(),
		NewNoIndexInSitemapReporter(),
		NewSitemapAndBlockedReporter(),
		NewNonCanonicalInSitemapReporter(),
		NewCanonicalMultipleTagsReporter(),
		NewCanonicalRelativeURLReporter(),
		NewCanonicalMismatchReporter(),
		NewDepthReporter(),
		NewNosnippetReporter(),
		NewMetasInBodyReporter(),

		// Add link issue reporters
		NewTooManyLinksReporter(),
		NewInternalNoFollowLinksReporter(),
		NewExternalLinkWitoutNoFollowReporter(),
		NewHTTPLinksReporter(),
		NewDeadendReporter(),
		NewExternalLinkRedirectReporter(),
		NewExternalLinkBrokenReporter(),
		NewLocalhostLinksReporter(),

		// Add image issue reporters
		NewAltTextReporter(),
		NewLongAltTextReporter(),
		NewLargeImageReporter(),
		NewNoImageIndexReporter(),
		NewMissingImgTagInPictureReporter(),
		NewImgWithoutSizeReporter(),

		// Add language issue reporters
		NewInvalidLangReporter(),
		NewMissingLangReporter(),
		NewHreflangXDefaultMissingReporter(),
		NewHreflangMissingSelfReference(),
		NewHreflangMismatchingLang(),
		NewHreflangRelativeURL(),

		// Add heading issue reporters
		NewNoH1Reporter(),
		NewValidHeadingsOrderReporter(),

		// Add content issue reporters
		NewLittleContentReporter(),
		NewIncorrectMediaTypeReporter(),
		NewDuplicatedIdReporter(),
		NewDOMSizeReporter(1500), // report html documents with more than 1500 nodes
		NewPaginationReporter(),

		// Add scheme issue reporters
		NewHTTPSchemeReporter(),

		// Add security issue reporters
		NewMissingHSTSHeaderReporter(),
		NewMissingCSPReporter(),
		NewMissingContentTypeOptionsReporter(),

		// Add timeout issue reporter
		NewTimeoutReporter(),

		// Add URL issue reports
		NewUnderscoreURLReporter(),
		NewSpaceURLReporter(),
		NewMultipleSlashesReporter(),

		// Add Time To Firts Byte reporter
		NewSlowTTFBReporter(),

		// Add form reporters
		NewFormOnHTTPReporter(),
		NewInsecureFormReporter(),

		// Add Viewport issue report
		NewViewportTagReporter(),
	}
}
