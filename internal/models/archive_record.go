package models

import "net/http"

type ArchiveRecord struct {
	Headers http.Header
	Body    string
}
