package reporters

import (
	"github.com/stjudewashere/seonaut/internal/models"
)

const (
	Error30x                         = iota + 1 // HTTP redirect
	Error40x                                    // HTTP not found
	Error50x                                    // HTTP internal error
	ErrorDuplicatedTitle                        // Duplicate title
	ErrorDuplicatedDescription                  // Duplicate description
	ErrorEmptyTitle                             // Missing or empty title
	ErrorShortTitle                             // Page title is too short
	ErrorLongTitle                              // Page title is too long
	ErrorEmptyDescription                       // Missing or empty meta description
	ErrorShortDescription                       // Meta description is too short
	ErrorLongDescription                        // Meta description is too long
	ErrorLittleContent                          // Not enough content
	ErrorImagesWithNoAlt                        // Images with no alt attribute
	ErrorRedirectChain                          // Redirect chain
	ErrorNoH1                                   // Missing or empy H1 tag
	ErrorNoLang                                 // Missing or empty html lang attribute
	ErrorHTTPLinks                              // Links using insecure http schema
	ErrorHreflangsReturnLink                    // Hreflang is not bidirectional
	ErrorTooManyLinks                           // Page contains too many links
	ErrorInternalNoFollow                       // Page has internal links with nofollow attribute
	ErrorExternalWithoutNoFollow                // Page has external follow links
	ErrorCanonicalizedToNonCanonical            // Page canonicalized to a non canonical page
	ErrorRedirectLoop                           // Redirect loop
	ErrorNotValidHeadings                       // H1-H6 tags have wrong order
	HreflangToNonCanonical                      // Hreflang to non canonical page
	ErrorInternalNoFollowIndexable              // Nofollow links to indexable pages
	ErrorNoIndexable                            // Page using the noindex attribute
	ErrorHreflangNoindexable                    // Hreflang to a non indexable page
	ErrorBlocked                                // Blocked by robots.txt
	ErrorOrphan                                 // Orphan pages
	SitemapNoIndex                              // No index pages included in the sitemap
	SitemapBlocked                              // Pages included in the sitemap that are blocked in robots.txt
	SitemapNonCanonical                         // Non canonical pages included in the sitemap
	IncomingFollowNofollow                      // Pages with index and noindex incoming links
	InvalidLanguage                             // Pages with invalid lang attribute
)

type MultipageCallback func(c *models.Crawl) *MultipageIssueReporter

type PageIssueReporter struct {
	Callback  func(*models.PageReport) bool
	ErrorType int
}

type MultipageIssueReporter struct {
	Query      string
	Parameters []interface{}
	ErrorType  int
}
