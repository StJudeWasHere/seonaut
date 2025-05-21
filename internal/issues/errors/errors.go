// Package reporter_errors defines constants for the different types of errors
// that can be found by the reporters.
//
// This package contains constants for different error types, where each constant
// has a unique integer ID that matches the corresponding error's ID in the database.
// These constants help identify and handle specific types of errors in the code.
// Additionally, reporters use these constants to report the type of issue they have
// found in the crawled pages.
package errors

const (
	Error30x                          = iota + 1 // HTTP redirect
	Error40x                                     // HTTP not found
	Error50x                                     // HTTP internal error
	ErrorDuplicatedTitle                         // Duplicate title
	ErrorDuplicatedDescription                   // Duplicate description
	ErrorEmptyTitle                              // Missing or empty title
	ErrorShortTitle                              // Page title is too short
	ErrorLongTitle                               // Page title is too long
	ErrorEmptyDescription                        // Missing or empty meta description
	ErrorShortDescription                        // Meta description is too short
	ErrorLongDescription                         // Meta description is too long
	ErrorLittleContent                           // Not enough content
	ErrorImagesWithNoAlt                         // Images with no alt attribute
	ErrorRedirectChain                           // Redirect chain
	ErrorNoH1                                    // Missing or empy H1 tag
	ErrorNoLang                                  // Missing or empty html lang attribute
	ErrorHTTPLinks                               // Links using insecure http schema
	ErrorHreflangsReturnLink                     // Hreflang is not bidirectional
	ErrorTooManyLinks                            // Page contains too many links
	ErrorInternalNoFollow                        // Page has internal links with nofollow attribute
	ErrorExternalWithoutNoFollow                 // Page has external follow links
	ErrorCanonicalizedToNonCanonical             // Page canonicalized to a non canonical page
	ErrorRedirectLoop                            // Redirect loop
	ErrorNotValidHeadings                        // H1-H6 tags have wrong order
	ErrorHreflangToNonCanonical                  // Hreflang to non canonical page
	ErrorInternalNoFollowIndexable               // Nofollow links to indexable pages
	ErrorNoIndexable                             // Page using the noindex attribute
	ErrorHreflangNoindexable                     // Hreflang to a non indexable page
	ErrorBlocked                                 // Blocked by robots.txt
	ErrorOrphan                                  // Orphan pages
	ErrorSitemapNoIndex                          // No index pages included in the sitemap
	ErrorSitemapBlocked                          // Pages included in the sitemap that are blocked in robots.txt
	ErrorSitemapNonCanonical                     // Non canonical pages included in the sitemap
	ErrorIncomingFollowNofollow                  // Pages with follow and nofollow incoming links
	ErrorInvalidLanguage                         // Pages with invalid lang attribute
	ErrorHTTPScheme                              // Pages using http scheme instead of https
	ErrorDeadend                                 // Pages with no outgoing internal or external links
	ErrorCanonicalizedToNonIndexable             // Pages that are canonicalized to non-indexable pages
	ErrorHreflangToRedirect                      // Pages that have hreflang links to other redirected pages
	ErrorCanonicalizedToRedirect                 // Pages that are canonicalized to other redirected pages
	ErrorHreflangToError                         // Pages that have hreflang links to error pages
	ErrorCanonicalizedToError                    // Pages that are canonicalized to error pages
	ErrorMultipleCanonicalTags                   // Pages that have more than one canonical tag
	ErrorRelativeCanonicalURL                    // Pages that are using a relative canonical URL
	ErrorHreflangMissingXDefault                 // Pages with hreflang tags and missing x-default value
	ErrorHreflangMissingSelfReference            // Pages with hreflang tags and missing self-reference
	ErrorHreflangMismatchLang                    // Pages with hreflang and mismatching lang in self-reference
	ErrorHreflangRelativeURL                     // Pages using relative urls hreflang links
	ErrorCanonicalMismatch                       // Pages with different canonical URLs in the HTML and HTTP headers
	ErrorMissingHSTSHeader                       // Pages with missing HSTS header
	ErrorMissingCSP                              // Pages with missing content security policy
	ErrorContentTypeOptions                      // Pages missing the X-Content-Type-Options header
	ErrorLargeImage                              // Large images
	ErrorLongAltText                             // Pages with images that have a long alt text
	ErrorMultipleTitleTags                       // Pages with more than one title tag in the header
	ErrorMultipleDescriptionTags                 // Pages with more than one meta description tag
	ErrorDepth                                   // Pages with high depth
	ErrorMultipleLangReference                   // Pages referenced with multiple languages in hreflangs
	ErrorDuplicatedContent                       // Pages with the same exact content
	ErrorExternalLinkRedirect                    // Pages with external links to redirect URLs
	ErrorExternalLinkBroken                      // Pages with brooken external links
	ErrorTimeout                                 // Pages that timed out
	ErrorUnderscoreURL                           // Pages wich URL has underscore characters
	ErrorSlowTTFB                                // Pages with slow time to first byte
	ErrorFormOnHTTP                              // Pages with forms on HTTP URLs
	ErrorInsecureForm                            // Forms with HTTP action URLs
	ErrorSpaceURL                                // URLS containing spaces
	ErrorMultipleSlashes                         // URLS containing multiple slashes in their path
	ErrorNoImageIndex                            // Pages with the noimageindex rule in the robots meta
	ErrorMissingImgElement                       // Pages with Picture missing the img element
	ErrorMetasInBody                             // Pages with meta tags in the document's body
	ErrorNosnippet                               // Pages with the nosnippet directive
	ErrorImgWithoutSize                          // Pages with img elements that have no size attribtues
	ErrorIncorrectMediaType                      // URLs with incorrect media type or media type that doesn't match extension
	ErrorDuplicatedId                            // Pages with duplicated id attributes
	ErrorMissingViewportTag                      // Pages with missing viewport meta tag
	ErrorDOMSize                                 // HTML documents with excessive DOM size
	ErrorPaginationLink                          // Pages with next and prev attributes missing the actual link
	ErrorLocalhostLinks                          // Pages with links to localhost or 127.0.0.1
)
