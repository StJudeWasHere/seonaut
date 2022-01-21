package main

import (
	"fmt"
)

const (
	Error30x                   = "ERROR_30x"
	Error40x                   = "ERROR_40x"
	Error50x                   = "ERROR_50x"
	ErrorDuplicatedTitle       = "ERROR_DUPLICATED_TITLE"
	ErrorDuplicatedDescription = "ERROR_DUPLICATED_DESCRIPTION"
	ErrorEmptyTitle            = "ERROR_EMPTY_TITLE"
	ErrorShortTitle            = "ERROR_SHORT_TITLE"
	ErrorLongTitle             = "ERROR_LONG_TITLE"
	ErrorEmptyDescription      = "ERROR_EMPTY_DESCRIPTION"
	ErrorShortDescription      = "ERROR_SHORT_DESCRIPTION"
	ErrorLongDescription       = "ERROR_LONG_DESCRIPTION"
	ErrorLittleContent         = "ERROR_LITTLE_CONTENT"
	ErrorImagesWithNoAlt       = "ERROR_IMAGES_NO_ALT"
	ErrorRedirectChain         = "ERROR_REDIRECT_CHAIN"
)

type Issue struct {
	PageReportId int
	ErrorType    string
}

type IssueGroup struct {
	ErrorType string
	Count     int
}

type IssueCallback struct {
	Callback  func(int) []PageReport
	ErrorType string
}

func createIssues(cid int) {
	var issues []Issue
	fmt.Println("Creating issues...")

	callbacks := []IssueCallback{
		IssueCallback{
			Callback:  Find30xPageReports,
			ErrorType: Error30x,
		},
		IssueCallback{
			Callback:  Find40xPageReports,
			ErrorType: Error30x,
		},
		IssueCallback{
			Callback:  Find50xPageReports,
			ErrorType: Error30x,
		},
		IssueCallback{
			Callback:  FindPageReportsWithDuplicatedTitle,
			ErrorType: ErrorDuplicatedTitle,
		},
		IssueCallback{
			Callback:  FindPageReportsWithDuplicatedTitle,
			ErrorType: ErrorDuplicatedDescription,
		},
		IssueCallback{
			Callback:  FindPageReportsWithEmptyTitle,
			ErrorType: ErrorEmptyTitle,
		},
		IssueCallback{
			Callback:  FindPageReportsWithShortTitle,
			ErrorType: ErrorShortTitle,
		},
		IssueCallback{
			Callback:  FindPageReportsWithLongTitle,
			ErrorType: ErrorLongTitle,
		},
		IssueCallback{
			Callback:  FindPageReportsWithEmptyDescription,
			ErrorType: ErrorEmptyDescription,
		},
		IssueCallback{
			Callback:  FindPageReportsWithShortDescription,
			ErrorType: ErrorShortDescription,
		},
		IssueCallback{
			Callback:  FindPageReportsWithLongDescription,
			ErrorType: ErrorLongDescription,
		},
		IssueCallback{
			Callback:  FindPageReportsWithLittleContent,
			ErrorType: ErrorLittleContent,
		},
		IssueCallback{
			Callback:  FindImagesWithNoAlt,
			ErrorType: ErrorImagesWithNoAlt,
		},
		IssueCallback{
			Callback:  FindImagesWithNoAlt,
			ErrorType: ErrorRedirectChain,
		},
	}

	for _, c := range callbacks {
		for _, p := range c.Callback(cid) {
			i := Issue{
				PageReportId: p.Id,
				ErrorType:    c.ErrorType,
			}

			issues = append(issues, i)
		}
	}

	saveIssues(issues, cid)
}
