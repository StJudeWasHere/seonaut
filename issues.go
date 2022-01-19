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

	LevelFatal   = 1
	LevelError   = 2
	LevelWarning = 3
)

type Issue struct {
	PageReportId int
	ErrorType    string
	Level        int
}

type IssueGroup struct {
	ErrorType string
	Level     int
	Count     int
}

type IssueCallback struct {
	Callback  func(int) []PageReport
	ErrorType string
	Level     int
}

func createIssues(cid int) {
	var issues []Issue
	fmt.Println("Creating issues...")

	callbacks := []IssueCallback{
		IssueCallback{
			Callback:  Find30xPageReports,
			ErrorType: Error30x,
			Level:     LevelFatal,
		},
		IssueCallback{
			Callback:  Find40xPageReports,
			ErrorType: Error30x,
			Level:     LevelFatal,
		},
		IssueCallback{
			Callback:  Find50xPageReports,
			ErrorType: Error30x,
			Level:     LevelFatal,
		},
		IssueCallback{
			Callback:  FindPageReportsWithDuplicatedTitle,
			ErrorType: ErrorDuplicatedTitle,
			Level:     LevelError,
		},
		IssueCallback{
			Callback:  FindPageReportsWithDuplicatedTitle,
			ErrorType: ErrorDuplicatedDescription,
			Level:     LevelError,
		},
		IssueCallback{
			Callback:  FindPageReportsWithEmptyTitle,
			ErrorType: ErrorEmptyTitle,
			Level:     LevelError,
		},
		IssueCallback{
			Callback:  FindPageReportsWithShortTitle,
			ErrorType: ErrorShortTitle,
			Level:     LevelWarning,
		},
		IssueCallback{
			Callback:  FindPageReportsWithLongTitle,
			ErrorType: ErrorLongTitle,
			Level:     LevelWarning,
		},
		IssueCallback{
			Callback:  FindPageReportsWithEmptyDescription,
			ErrorType: ErrorEmptyDescription,
			Level:     LevelWarning,
		},
		IssueCallback{
			Callback:  FindPageReportsWithShortDescription,
			ErrorType: ErrorShortDescription,
			Level:     LevelWarning,
		},
		IssueCallback{
			Callback:  FindPageReportsWithLongDescription,
			ErrorType: ErrorLongDescription,
			Level:     LevelWarning,
		},
		IssueCallback{
			Callback:  FindPageReportsWithLittleContent,
			ErrorType: ErrorLittleContent,
			Level:     LevelWarning,
		},
		IssueCallback{
			Callback:  FindImagesWithNoAlt,
			ErrorType: ErrorImagesWithNoAlt,
			Level:     LevelWarning,
		},
	}

	for _, c := range callbacks {
		for _, p := range c.Callback(cid) {
			i := Issue{
				PageReportId: p.Id,
				ErrorType:    c.ErrorType,
				Level:        c.Level,
			}

			issues = append(issues, i)
		}
	}

	saveIssues(issues, cid)
}
