package services_test

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/stjudewashere/seonaut/internal/services"
)

const (
	testURL = "https://example.com/test-page/"
)

func TestNewPageReport(t *testing.T) {
	u, err := url.Parse(testURL)
	if err != nil {
		fmt.Println(err)
	}

	contentType := "text/html"
	statusCode := 200
	body := []byte("<html>")

	headers := http.Header{
		"Content-Type": []string{contentType},
	}

	pageReport, _, err := services.NewHTMLParser(u, statusCode, &headers, body, int64(len(body)))
	if err != nil {
		t.Fatal(err)
	}

	if pageReport.URL != testURL {
		t.Error("NewPageReport URL != testURL")
	}

	if pageReport.ParsedURL != u {
		t.Error("NewPageReport ParsedURL != u")
	}

	if pageReport.StatusCode != statusCode {
		t.Error("NewPageReport StatusCode != statusCode")
	}

	if pageReport.ContentType != "text/html" {
		t.Error("NewPageReport ContentType != contentType")
	}
}

func TestNewRedirectPageReport(t *testing.T) {
	u, err := url.Parse(testURL)
	if err != nil {
		fmt.Println(err)
	}

	body := []byte("<html>")
	statusCode := 301
	redirectURL := "https://example.com/redirect"

	headers := http.Header{
		"Location":     []string{redirectURL},
		"Content-Type": []string{"text/html"},
	}

	pageReport, _, err := services.NewHTMLParser(u, statusCode, &headers, body, int64(len(body)))
	if err != nil {
		t.Fatal(err)
	}

	if pageReport.RedirectURL != redirectURL {
		t.Errorf("NewPageReport RedirectURL != %s", pageReport.RedirectURL)
	}

	if pageReport.StatusCode != statusCode {
		t.Error("NewPageReport StatusCode != statusCode")
	}
}

func TestPageReportHTML(t *testing.T) {
	u, err := url.Parse(testURL)
	if err != nil {
		fmt.Println(err)
	}

	contentType := "text/html"
	statusCode := 200
	headers := &http.Header{
		"Content-Type": []string{contentType},
	}
	body, err := os.ReadFile("./testdata/parser.html")
	if err != nil {
		log.Fatal(err)
	}

	pageReport, _, err := services.NewHTMLParser(u, statusCode, headers, body, int64(len(body)))
	if err != nil {
		t.Fatal(err)
	}

	itable := []struct {
		want int
		got  int
	}{
		{want: 6, got: len(pageReport.Links)},
		{want: 1, got: len(pageReport.ExternalLinks)},
		{want: 10, got: pageReport.Words},
		{want: 2, got: len(pageReport.Hreflangs)},
		{want: 7, got: len(pageReport.Images)},
		{want: 1, got: len(pageReport.Scripts)},
		{want: 1, got: len(pageReport.Styles)},
		{want: 1, got: len(pageReport.Iframes)},
		{want: 3, got: len(pageReport.Audios)},
		{want: 3, got: len(pageReport.Videos)},
	}

	stable := []struct {
		want string
		got  string
	}{
		{want: "https://example.com/fr", got: pageReport.Hreflangs[0].URL},
		{want: "fr", got: pageReport.Hreflangs[0].Lang},
		{want: "https://example.com/js/app.js", got: pageReport.Scripts[0]},
		{want: "https://example.com/css/style.css", got: pageReport.Styles[0]},
		{want: "en", got: pageReport.Lang},
		{want: "Test Page Title", got: pageReport.Title},
		{want: "Test Page Description", got: pageReport.Description},
		{want: "https://example.com/link1", got: pageReport.Links[0].URL},
		{want: "https://example.com/test-page/link2", got: pageReport.Links[1].URL},
		{want: "link1", got: pageReport.Links[0].Text},
		{want: "nofollow", got: pageReport.Links[0].Rel},
		{want: "", got: pageReport.Links[3].Text},
		{want: "https://example.com/", got: pageReport.Links[4].URL},
		{want: "https://example.com/test-page/", got: pageReport.Links[5].URL},
		{want: "0;URL='/'", got: pageReport.Refresh},
		{want: "https://example.com/", got: pageReport.RedirectURL},
		{want: "noindex, nofollow", got: pageReport.Robots},
		{want: "https://example.com/canonical/", got: pageReport.Canonical},
		{want: "H1 Title", got: pageReport.H1},
		{want: "H2 Title", got: pageReport.H2},
		{want: "https://example.com/img/logo.png", got: pageReport.Images[0].URL},
		{want: "http://example.com/", got: pageReport.Iframes[0]},
		{want: "https://example.com/audio_file.ogg", got: pageReport.Audios[0]},
		{want: "https://example.com/audio_file.wav", got: pageReport.Audios[1]},
		{want: "https://example.com/audio_file.mp3", got: pageReport.Audios[2]},
		{want: "https://example.com/video_file.webm", got: pageReport.Videos[0].URL},
		{want: "https://example.com/poster.png", got: pageReport.Videos[0].Poster},
		{want: "https://example.com/poster.png", got: pageReport.Videos[1].Poster},
		{want: "", got: pageReport.Videos[2].Poster},
	}

	btable := []struct {
		want bool
		got  bool
	}{
		{want: false, got: pageReport.Links[0].External},
		{want: true, got: pageReport.Noindex},
		{want: true, got: pageReport.ExternalLinks[0].Sponsored},
		{want: true, got: pageReport.ExternalLinks[0].UGC},
	}

	for _, v := range itable {
		if v.want != v.got {
			t.Errorf("want: %d got: %d", v.want, v.got)
		}
	}

	for _, v := range stable {
		if v.got != v.want {
			t.Errorf("want: %s got: %s", v.want, v.got)
		}
	}

	for _, v := range btable {
		if v.got != v.want {
			t.Errorf("want: %v got: %v", v.want, v.got)
		}
	}
}

func TestMultipleCanonicalTags(t *testing.T) {
	u, err := url.Parse(testURL)
	if err != nil {
		fmt.Println(err)
	}

	statusCode := 200
	headers := http.Header{
		"Content-Type": []string{"text/html"},
	}
	body := []byte(`
		<html>
			<head>
				<link rel="canonical" href="/canonical-1/" />
				<link rel="canonical" href="/canonical-2/" />
			</head>
		`)

	pageReport, _, err := services.NewHTMLParser(u, statusCode, &headers, body, int64(len(body)))
	if err != nil {
		t.Fatal(err)
	}

	if pageReport.Canonical != "" {
		t.Error("Multiple canonical tags should be ignored ")
	}
}

func TestCanonicalTagInBody(t *testing.T) {
	u, err := url.Parse(testURL)
	if err != nil {
		fmt.Println(err)
	}

	statusCode := 200
	headers := http.Header{
		"Content-Type": []string{"text/html"},
	}
	body := []byte(`
		<html>
			<head></head>
			<body>
				<link rel="canonical" href="/canonical-1/" />
			</body>
		`)

	pageReport, _, err := services.NewHTMLParser(u, statusCode, &headers, body, int64(len(body)))
	if err != nil {
		t.Fatal(err)
	}

	if pageReport.Canonical != "" {
		t.Error("Canonical tags in body should be ignored ")
	}
}

func TestNoindex(t *testing.T) {
	u, err := url.Parse(testURL)
	if err != nil {
		fmt.Println(err)
	}

	body := []byte("<html>")
	statusCode := 200
	headers := http.Header{
		"X-Robots-Tag": []string{"noindex, nofollow"},
		"Content-Type": []string{"text/html"},
	}

	pageReport, _, err := services.NewHTMLParser(u, statusCode, &headers, body, int64(len(body)))
	if err != nil {
		t.Fatal(err)
	}

	if pageReport.Nofollow == false {
		t.Error("Nofollow == false")
	}

	if pageReport.Noindex == false {
		t.Error("Noindex == false")
	}
}

func TestContentLanguage(t *testing.T) {
	u, err := url.Parse(testURL)
	if err != nil {
		fmt.Println(err)
	}

	body := []byte("<html>")
	statusCode := 200
	contentLanguage := "en-us"
	headers := http.Header{
		"Content-Language": []string{contentLanguage},
		"Content-Type":     []string{"text/html"},
	}

	pageReport, _, err := services.NewHTMLParser(u, statusCode, &headers, body, int64(len(body)))
	if err != nil {
		t.Fatal(err)
	}

	if pageReport.Lang != contentLanguage {
		t.Errorf("ContentLanguage: %s != %s", pageReport.Lang, contentLanguage)
	}
}

func TestHreflangHeaders(t *testing.T) {
	u, err := url.Parse(testURL)
	if err != nil {
		fmt.Println(err)
	}

	linkHeader := `
		<https://example.com/file.pdf>; rel="alternate"; hreflang="en",
		<https://de-ch.example.com/file.pdf>; rel="alternate"; hreflang="de-ch",
		<https://de.example.com/file.pdf>; rel="alternate"; hreflang="de"
	`

	body := []byte("<html>")
	statusCode := 200
	headers := http.Header{
		"Link":         []string{linkHeader},
		"Content-Type": []string{"text/html"},
	}

	pageReport, _, err := services.NewHTMLParser(u, statusCode, &headers, body, int64(len(body)))
	if err != nil {
		t.Fatal(err)
	}

	if len(pageReport.Hreflangs) != 3 {
		t.Errorf("HreflangHeader: %d != 3", len(pageReport.Hreflangs))
	}

	if pageReport.Hreflangs[0].URL != "https://example.com/file.pdf" || pageReport.Hreflangs[0].Lang != "en" {
		t.Errorf("HreflangHeader: Hreflangs[0]: %v ", pageReport.Hreflangs[0])
	}

	if pageReport.Hreflangs[1].URL != "https://de-ch.example.com/file.pdf" || pageReport.Hreflangs[1].Lang != "de-ch" {
		t.Errorf("HreflangHeader: Hreflangs[1]: %v ", pageReport.Hreflangs[1])
	}

	if pageReport.Hreflangs[2].URL != "https://de.example.com/file.pdf" || pageReport.Hreflangs[2].Lang != "de" {
		t.Errorf("HreflangHeader: Hreflangs[2]: %v ", pageReport.Hreflangs[2])
	}
}

func TestCanonicalHeaders(t *testing.T) {
	u, err := url.Parse(testURL)
	if err != nil {
		fmt.Println(err)
	}

	linkHeader := `
		<https://example.com/canonical>; rel="canonical",
		<https://de-ch.example.com/file.pdf>; rel="alternate"; hreflang="de-ch",
		<https://de.example.com/file.pdf>; rel="alternate"; hreflang="de"
	`

	body := []byte("<html>")
	statusCode := 200
	headers := http.Header{
		"Link":         []string{linkHeader},
		"Content-Type": []string{"text/html"},
	}

	pageReport, _, err := services.NewHTMLParser(u, statusCode, &headers, body, int64(len(body)))
	if err != nil {
		t.Fatal(err)
	}

	if pageReport.Canonical != "https://example.com/canonical" {
		t.Errorf("Canonical headers: %s != https://example.com/canonical", pageReport.Canonical)
	}
}

func TestRelativeCanonicalHeaders(t *testing.T) {
	u, err := url.Parse(testURL)
	if err != nil {
		fmt.Println(err)
	}

	linkHeader := `
		</canonical>; rel="canonical",
		<https://de-ch.example.com/file.pdf>; rel="alternate"; hreflang="de-ch",
		<https://de.example.com/file.pdf>; rel="alternate"; hreflang="de"
	`

	body := []byte("<html>")
	statusCode := 200
	headers := http.Header{
		"Link":         []string{linkHeader},
		"Content-Type": []string{"text/html"},
	}

	pageReport, _, err := services.NewHTMLParser(u, statusCode, &headers, body, int64(len(body)))
	if err != nil {
		t.Fatal(err)
	}

	if pageReport.Canonical != "https://example.com/canonical" {
		t.Errorf("Canonical headers: %s != https://example.com/canonical", pageReport.Canonical)
	}
}

func TestNoBodyTag(t *testing.T) {
	u, err := url.Parse(testURL)
	if err != nil {
		fmt.Println(err)
	}

	body := []byte("<html><frameset></frameset></html>")
	statusCode := 200
	headers := http.Header{
		"Content-Type": []string{"text/html"},
	}

	pageReport, _, err := services.NewHTMLParser(u, statusCode, &headers, body, int64(len(body)))
	if err != nil {
		t.Fatal(err)
	}

	if pageReport.Words != 0 {
		t.Errorf("NoBody: %d != 0", pageReport.Words)
	}
}

func TestRobotsNone(t *testing.T) {
	u, err := url.Parse(testURL)
	if err != nil {
		fmt.Println(err)
	}

	body := []byte(
		`<html>
		<head><meta name="robots" content="none"></head>
		<body></body>
	</html>`)
	statusCode := 200
	headers := &http.Header{
		"Content-Type": []string{"text/html"},
	}

	pageReport, _, err := services.NewHTMLParser(u, statusCode, headers, body, int64(len(body)))
	if err != nil {
		t.Fatal(err)
	}

	if !pageReport.Noindex {
		t.Errorf("NewPageReport Noindex should be true")
	}

	if !pageReport.Nofollow {
		t.Error("NewPageReport Nofollow should be true")
	}
}

func TestSrcset(t *testing.T) {
	u, err := url.Parse(testURL)
	if err != nil {
		fmt.Println(err)
	}

	images := []string{
		"https://example.com/logo.png",
		"https://example.com/image,c_fill,w_576.jpg",
		"https://example.com/image,c_fill,w_276.jpg",
		"https://example.com/image,c_fill,w_76.jpg",
	}
	body := []byte(
		`<html>
		<head></head>
		<body>
			<img src="` + images[0] + `"
			srcset=",` + images[1] + ` 576w, ,` + images[2] + ` 276w,` + images[3] + `,">
		</body>
	</html>`)
	statusCode := 200
	headers := &http.Header{
		"Content-Type": []string{"text/html"},
	}

	pageReport, _, err := services.NewHTMLParser(u, statusCode, headers, body, int64(len(body)))
	if err != nil {
		t.Fatal(err)
	}

	if len(pageReport.Images) != len(images) {
		t.Errorf("pagereport images len want: %d Got: %d", len(images), len(pageReport.Images))
	}

	for n, i := range images {
		if pageReport.Images[n].URL != i {
			t.Errorf("pageReport image %d should be %s. Got: %s", n, i, pageReport.Images[n].URL)
		}
	}
}

func TestEmptyBody(t *testing.T) {
	u, err := url.Parse(testURL)
	if err != nil {
		fmt.Println(err)
	}

	body := []byte("")
	statusCode := 404
	headers := &http.Header{
		"Content-Type": []string{"text/html"},
	}

	pageReport, _, err := services.NewHTMLParser(u, statusCode, headers, body, int64(len(body)))
	if err != nil {
		t.Fatal(err)
	}

	if pageReport.StatusCode != statusCode {
		t.Errorf("pageReport status code should be %d but received %d", statusCode, pageReport.StatusCode)
	}
}

func TestBase(t *testing.T) {
	u, err := url.Parse("https://example-base.com")
	if err != nil {
		fmt.Println(err)
	}

	body := []byte(
		`<html>
		<head><base href="https://example-base.com/test"></head>
		<body>
			<a href="/page.html">link</a>
			<a href="../category/page.html">link 2</a>
			<a href="https://external.com/page.html">link 2</a>
		</body>
	</html>`)
	statusCode := 200
	headers := &http.Header{
		"Content-Type": []string{"text/html"},
	}

	pageReport, _, err := services.NewHTMLParser(u, statusCode, headers, body, int64(len(body)))
	if err != nil {
		t.Fatal(err)
	}

	if len(pageReport.Links) != 2 {
		t.Fatal("wrong number of links with base URL")
	}

	if pageReport.Links[0].URL != "https://example-base.com/test/page.html" {
		t.Errorf("Link with base URL does not match, got %s", pageReport.Links[0].URL)
	}

	if pageReport.Links[1].URL != "https://example-base.com/category/page.html" {
		t.Errorf("Link with base URL does not match, got %s", pageReport.Links[1].URL)
	}

	if len(pageReport.ExternalLinks) != 1 {
		t.Fatal("An external link with base URL was expected")
	}

	if pageReport.ExternalLinks[0].URL != "https://external.com/page.html" {
		t.Errorf("External link with base URL does not match, got %s", pageReport.ExternalLinks[0].URL)
	}
}

func TestBaseRelativeURL(t *testing.T) {
	u, err := url.Parse(testURL)
	if err != nil {
		fmt.Println(err)
	}

	body := []byte(
		`<html>
		<head><base href="/test"></head>
		<body>
			<a href="/page.html">link</a>
			<a href="https://example.com/page.html">link</a>
			<a href="https://example.com/../category/page.html">link</a>
			<a href="https://external.com/category/page.html">link</a>
		</body>
	</html>`)
	statusCode := 200
	headers := &http.Header{
		"Content-Type": []string{"text/html"},
	}

	pageReport, _, err := services.NewHTMLParser(u, statusCode, headers, body, int64(len(body)))
	if err != nil {
		t.Fatal(err)
	}

	table := []struct {
		want string
		got  string
	}{
		{"https://example.com/test/page.html", pageReport.Links[0].URL},
		{"https://example.com/page.html", pageReport.Links[1].URL},
		{"https://example.com/../category/page.html", pageReport.Links[2].URL},
	}

	if len(pageReport.Links) != len(table) {
		t.Errorf("expected %d links but got %d", len(table), len(pageReport.Links))
	}

	for _, n := range table {
		if n.got != n.want {
			t.Errorf("Link with base URL %s does not match, got %s", n.want, n.got)
		}
	}

	if len(pageReport.ExternalLinks) != 1 {
		t.Fatal("Number of expected external links don't match")
	}

	if pageReport.ExternalLinks[0].URL != "https://external.com/category/page.html" {
		t.Errorf("Link with base URL does not match, got %s", pageReport.ExternalLinks[0].URL)
	}
}
