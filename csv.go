package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"unicode/utf8"
)

var writer *csv.Writer

func init() {
	f, e := os.Create("./seo.csv")
	if e != nil {
		fmt.Println(e)
	}

	writer = csv.NewWriter(f)

	writer.Write([]string{
		"Status Code",
		"URL",
		"Redirect URL",
		"Content Type",
		"Canonical",
		"Lang",
		"Title",
		"Title Length",
		"Description",
		"Description Length",
		"Robots",
		"Header 1",
		"Header 2",
		"Size",
		"Nº of words",
		"Internal links",
		"External links",
	})
}

func handlePageReport(r PageReport) {
	internal := make(map[string]bool)
	external := make(map[string]bool)

	for _, l := range r.Links {
		if l.External {
			external[l.URL] = true
		} else {
			internal[l.URL] = true
		}
	}

	writer.Write([]string{
		fmt.Sprintf("%d", r.StatusCode),
		r.URL.String(),
		r.RedirectURL,
		r.ContentType,
		r.Canonical,
		r.Lang,
		r.Title,
		fmt.Sprint(utf8.RuneCount([]byte(r.Title))),
		r.Description,
		fmt.Sprint(utf8.RuneCount([]byte(r.Description))),
		r.Robots,
		r.H1,
		r.H2,
		fmt.Sprintf("%.1f KB", byteToKByte(len(r.Body))),
		strconv.Itoa(r.Words),
		strconv.Itoa(len(internal)),
		strconv.Itoa(len(external)),
	})

	writer.Flush()
}

func byteToKByte(b int) float64 {
	v := b / (1 << 10)
	r := b % (1 << 10)

	return float64(v) + float64(r)/float64(1<<10)
}
