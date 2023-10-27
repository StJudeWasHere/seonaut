package reporters

import (
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/net/html"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager"
	"github.com/stjudewashere/seonaut/internal/report_manager/reporter_errors"
)

// Returns a report_manager.PageIssueReporter with a callback function that
// reports if the page's HSTS header is missing. The callback returns true if the Strict-Transport-Security,
// header does not exist or is not valid.
func NewMissingHSTSHeaderReporter() *report_manager.PageIssueReporter {
	c := func(pageReport *models.PageReport, htmlNode *html.Node, header *http.Header) bool {
		hstsHeader := header.Get("Strict-Transport-Security")
		if hstsHeader == "" {
			return true
		}

		directives := strings.Split(hstsHeader, ";")
		for _, directive := range directives {
			if strings.HasPrefix(directive, "max-age=") {
				maxAge := strings.TrimPrefix(directive, "max-age=")
				_, err := strconv.Atoi(maxAge)
				if err != nil {
					return true
				}
			}
		}

		return false
	}

	return &report_manager.PageIssueReporter{
		ErrorType: reporter_errors.ErrorMissingHSTSHeader,
		Callback:  c,
	}
}
