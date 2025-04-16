package models

type CanonicalCount struct {
	Canonical    int
	NonCanonical int
}

type SchemeCount struct {
	HTTP  int
	HTTPS int
}

type AltCount struct {
	Alt    int
	NonAlt int
}

type StatusCodeByDepth struct {
	Depth         int
	StatusCode100 int
	StatusCode200 int
	StatusCode300 int
	StatusCode400 int
	StatusCode500 int
}

