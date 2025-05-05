package models

type ExportLink struct {
	Origin      string
	Destination string
	Text        string
}

type ExportImage struct {
	Origin string
	Image  string
	Alt    string
}

type Script struct {
	Origin string
	Script string
}

type Style struct {
	Origin string
	Style  string
}

type Iframe struct {
	Origin string
	Iframe string
}

type Audio struct {
	Origin string
	Audio  string
}

type ExportVideo struct {
	Origin string
	Video  string
}

type ExportHreflang struct {
	Origin       string
	OriginLang   string
	Hreflang     string
	HreflangLang string
}

type ExportIssue struct {
	Url      string
	Type     string
	Priority int
}
