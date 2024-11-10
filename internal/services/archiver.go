package services

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/slyrz/warc"
)

const ArchiveDir = "archive/"

type Archiver struct {
	zipWriter    *zip.Writer
	file         *os.File
	cdxjEntries  []CDXJEntry
	pagesEntries []PageEntry
}

type CDXJEntry struct {
	TargetURI string
	Timestamp string
	RecordID  string
	Offset    string
	Status    string
	Length    string
	Mime      string
	Filename  string
	Digest    string
}

type PageEntry struct {
	URL string
	TS  string
}

// Returns a new Archiver.
// It creates a new wacz file for the given url string.
func NewArchiver(urlStr string) (*Archiver, error) {
	file, err := os.Create(ArchiveDir + urlStr + ".wacz")
	if err != nil {
		return nil, err
	}

	return &Archiver{
		zipWriter: zip.NewWriter(file),
		file:      file,
	}, nil
}

// AddRecord adds a new response record to the warc file and keeps track
// of the added records to create the index once the archiver is closed.
func (s *Archiver) AddRecord(response *http.Response) {
	uuidStr := uuid.New().String()
	record := warc.NewRecord()
	record.Header.Set("warc-type", "response")
	record.Header.Set("warc-date", time.Now().Format(time.RFC3339))
	record.Header.Set("warc-target-uri", response.Request.URL.String())
	record.Header.Set("content-type", response.Header.Get("Content-Type"))
	record.Header.Set("warc-record-id", fmt.Sprintf("<urn:uuid:%s>", uuidStr))

	var contentBuffer bytes.Buffer
	contentBuffer.WriteString(fmt.Sprintf("HTTP/%d.%d %d %s\r\n",
		response.ProtoMajor, response.ProtoMinor, response.StatusCode, response.Status))

	for key, values := range response.Header {
		for _, value := range values {
			contentBuffer.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
		}
	}
	contentBuffer.WriteString("\r\n")

	var bodyCopy bytes.Buffer
	_, err := io.Copy(&bodyCopy, response.Body)
	if err != nil {
		log.Printf("Failed to copy response body: %v", err)
		return
	}
	response.Body.Close() // Close the original body
	response.Body = io.NopCloser(bytes.NewReader(bodyCopy.Bytes()))

	if _, err := io.Copy(&contentBuffer, &bodyCopy); err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	record.Content = bytes.NewReader(contentBuffer.Bytes())

	filePath := ArchiveDir + fmt.Sprintf("data-%s.warc.gz", uuidStr)
	archiveFile, err := s.zipWriter.Create(filePath)
	if err != nil {
		log.Printf("Failed to create WARC file entry in ZIP: %v", err)
		return
	}
	archiveZipWritter := gzip.NewWriter(archiveFile)
	archiveWriter := warc.NewWriter(archiveZipWritter)

	if _, err := archiveWriter.WriteRecord(record); err != nil {
		log.Printf("Failed to write WARC record to archive: %v", err)
		return
	}

	archiveZipWritter.Close()

	cdxjEntry := CDXJEntry{
		TargetURI: response.Request.URL.String(),
		Timestamp: time.Now().Format("20060102150405"),
		RecordID:  fmt.Sprintf("<urn:uuid:%s>", uuidStr),
		Status:    fmt.Sprintf("%d", response.StatusCode),
		Length:    fmt.Sprintf("%d", len(contentBuffer.Bytes())),
		Mime:      response.Header.Get("Content-Type"),
		Filename:  filePath,
		Digest:    fmt.Sprintf("sha-256:%x", sha256.Sum256(contentBuffer.Bytes())),
		Offset:    fmt.Sprintf("%d", 0),
	}

	s.cdxjEntries = append(s.cdxjEntries, cdxjEntry)

	pageEntry := PageEntry{
		URL: response.Request.URL.String(),
		TS:  time.Now().Format(time.RFC3339),
	}

	s.pagesEntries = append(s.pagesEntries, pageEntry)
}

// Close closes the archive and creates the remaining files.
func (s *Archiver) Close() {
	pagesFile, err := s.zipWriter.Create("pages/pages.jsonl")
	if err != nil {
		log.Printf("Failed to create pages file entry in ZIP: %v", err)
		return
	}
	pagesWriter := gzip.NewWriter(pagesFile)

	header := `{"format": "json-pages-1.0", "id": "pages", "title": "All Pages"}`
	header += "\n"
	pagesWriter.Write([]byte(header))

	for _, page := range s.pagesEntries {
		pageLine := fmt.Sprintf(`{"url":"%s","ts":"%s"}`, page.URL, page.TS)
		pageLine += "\n"
		pagesWriter.Write([]byte(pageLine))
	}
	pagesWriter.Close()

	indexFile, err := s.zipWriter.Create("indexes/index.cdx.gz")
	if err != nil {
		log.Printf("Failed to create WARC file entry in ZIP: %v", err)
		return
	}
	indexWriter := gzip.NewWriter(indexFile)

	cdx := []string{}
	for _, entry := range s.cdxjEntries {
		parsedURL, err := url.Parse(entry.TargetURI)
		if err != nil {
			log.Printf("Failed to parse URL: %v", err)
			continue
		}
		domainParts := strings.Split(parsedURL.Hostname(), ".")
		slices.Reverse(domainParts)
		searchableURL := strings.Join(domainParts, ",")
		searchableURL = searchableURL + ")" + parsedURL.RequestURI()

		cdxjLine := fmt.Sprintf(
			"%s %s %s\n",
			searchableURL,
			entry.Timestamp,
			fmt.Sprintf(`{"offset":"%s","status":"%s","length":"%s","mime":"%s","filename":"%s","url":"%s","digest":"%s"}`,
				entry.Offset, entry.Status, entry.Length, entry.Mime, entry.Filename, entry.TargetURI, entry.Digest,
			),
		)
		cdx = append(cdx, cdxjLine)

	}
	slices.Sort(cdx)
	for _, e := range cdx {
		indexWriter.Write([]byte(e))
	}
	indexWriter.Close()

	s.zipWriter.Close()
	s.file.Close()

	err = s.createDatapackageJSON()
	if err != nil {
		log.Printf("Failed to create datapackage.json: %v", err)
		return
	}
}

// Create the datapackage.json file
func (s *Archiver) createDatapackageJSON() error {
	archive, err := zip.OpenReader(s.file.Name())
	if err != nil {
		log.Printf("Failed to open ZIP archive for reading: %v", err)
		return err
	}
	defer archive.Close()

	calculateSHA256AndSize := func(file *zip.File) (string, int64, error) {
		rc, err := file.Open()
		if err != nil {
			log.Printf("Failed to open file in ZIP: %v", err)
			return "", 0, err
		}
		defer rc.Close()

		hash := sha256.New()
		_, err = io.Copy(hash, rc)
		if err != nil {
			log.Printf("Failed to calculate SHA256: %v", err)
			return "", 0, err
		}

		fileSize := file.FileInfo().Size()

		return "sha256:" + hex.EncodeToString(hash.Sum(nil)), fileSize, nil
	}

	var resources []map[string]interface{}

	for _, file := range archive.File {
		hash, size, err := calculateSHA256AndSize(file)
		if err != nil {
			return err
		}

		resources = append(resources, map[string]interface{}{
			"name":  file.Name,
			"path":  file.Name,
			"hash":  hash,
			"bytes": size,
		})
	}

	datapackage := map[string]interface{}{
		"profile":      "data-package",
		"wacz_version": "1.1.1",
		"resources":    resources,
	}

	f, err := os.OpenFile(s.file.Name(), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
		return err
	}
	defer f.Close()

	zipWriter := zip.NewWriter(f)
	defer zipWriter.Close()

	for _, file := range archive.File {
		zipWriter.Copy(file)
	}

	datapackageFile, err := zipWriter.Create("datapackage.json")
	if err != nil {
		log.Printf("Failed to create datapackage.json in ZIP: %v", err)
		return err
	}

	encoder := json.NewEncoder(datapackageFile)
	err = encoder.Encode(datapackage)
	if err != nil {
		log.Printf("Failed to write datapackage.json: %v", err)
		return err
	}

	return nil
}
