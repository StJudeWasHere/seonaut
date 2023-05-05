// Package reporter_errors defines constants for the different types of errors
// that can be found by the reporters.
//
// This package contains constants for different error types, where each constant
// has a unique integer ID that matches the corresponding error's ID in the database.
// These constants help identify and handle specific types of errors in the code.
// Additionally, reporters use these constants to report the type of issue they have
// found in the crawled pages.
package reporter_errors

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
	ErrorHreflangToNonCanonical                 // Hreflang to non canonical page
	ErrorInternalNoFollowIndexable              // Nofollow links to indexable pages
	ErrorNoIndexable                            // Page using the noindex attribute
	ErrorHreflangNoindexable                    // Hreflang to a non indexable page
	ErrorBlocked                                // Blocked by robots.txt
	ErrorOrphan                                 // Orphan pages
	ErrorSitemapNoIndex                         // No index pages included in the sitemap
	ErrorSitemapBlocked                         // Pages included in the sitemap that are blocked in robots.txt
	ErrorSitemapNonCanonical                    // Non canonical pages included in the sitemap
	ErrorIncomingFollowNofollow                 // Pages with follow and nofollow incoming links
	ErrorInvalidLanguage                        // Pages with invalid lang attribute
	ErrorHTTPScheme                             // Pages using http scheme instead of https
)
