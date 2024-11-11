package services

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/slyrz/warc"
	"github.com/stjudewashere/seonaut/internal/models"
)

const ArchiveDir = "archive/"

type Archiver struct {
	file        *os.File
	cdxjEntries []CDXJEntry
	waczWriter  *zip.Writer
	warcWriter  *warc.Writer
	warcOffset  int
}

type CDXJEntry struct {
	Offset       string    `json:"offset"`
	Status       string    `json:"status"`
	Length       string    `json:"length"`
	Mime         string    `json:"mime"`
	Filename     string    `json:"filename"`
	Digest       string    `json:"digest"`
	RecordDigest string    `json:"recordDigest"`
	time         time.Time `json:"-"`
	url          url.URL   `json:"-"`
}

type PageEntry struct {
	URL string `json:"url"`
	TS  string `json:"ts"`
}

// Returns a new Archiver.
// It creates a new wacz file for the given url string.
func NewArchiver(p models.Project) (*Archiver, error) {
	// Create the project's archive directory if it doesn't exist.
	projectPath := ArchiveDir + "/" + strconv.FormatInt(p.Id, 10) + "/"
	err := os.MkdirAll(projectPath, 0755)
	if err != nil {
		return nil, err
	}

	// Create the wacz file.
	file, err := os.Create(projectPath + p.Host + ".wacz")
	if err != nil {
		return nil, err
	}
	waczWriter := zip.NewWriter(file)

	// Create the warc writer.
	archiveFile, err := waczWriter.Create("data/data.warc")
	if err != nil {
		log.Printf("failed to create WARC file entry in ZIP: %v", err)
		return nil, err
	}
	warcWriter := warc.NewWriter(archiveFile)

	return &Archiver{
		waczWriter: waczWriter,
		file:       file,
		warcWriter: warcWriter,
	}, nil
}

// AddRecord adds a new response record to the warc file and keeps track
// of the added records to create the index once the archiver is closed.
func (s *Archiver) AddRecord(response *http.Response) {
	var bodyCopy bytes.Buffer
	err := s.readResponseBody(&bodyCopy, response)
	if err != nil {
		log.Printf("failed to read response body %v", err)
		return
	}

	var contentBuffer bytes.Buffer
	s.readResponseHeaders(&contentBuffer, response)
	if _, err := io.Copy(&contentBuffer, &bodyCopy); err != nil {
		fmt.Println("error reading response body copy:", err)
		return
	}

	wdate := time.Now()
	record := warc.NewRecord()
	record.Header.Set("warc-type", "response")
	record.Header.Set("warc-date", wdate.Format(time.RFC3339))
	record.Header.Set("warc-target-uri", response.Request.URL.String())
	record.Header.Set("content-type", response.Header.Get("Content-Type"))
	record.Header.Set("warc-record-id", fmt.Sprintf("<urn:uuid:%s>", uuid.New().String()))
	record.Content = bytes.NewReader(contentBuffer.Bytes())
	if _, err := s.warcWriter.WriteRecord(record); err != nil {
		log.Printf("failed to write WARC record to archive: %v", err)
		return
	}

	cdxjEntry := CDXJEntry{
		Status:       fmt.Sprintf("%d", response.StatusCode),
		Length:       fmt.Sprintf("%d", len(contentBuffer.Bytes())),
		Mime:         response.Header.Get("Content-Type"),
		Filename:     "data/data.warc.gz",
		Digest:       fmt.Sprintf("sha-256:%x", sha256.Sum256(bodyCopy.Bytes())),
		RecordDigest: fmt.Sprintf("sha256:%x", sha256.Sum256(contentBuffer.Bytes())),
		Offset:       fmt.Sprintf("%d", s.warcOffset),
		time:         wdate,
		url:          *response.Request.URL,
	}
	s.cdxjEntries = append(s.cdxjEntries, cdxjEntry)

	s.warcOffset += contentBuffer.Len()
}

// readResponseBody Reads the http response's body into a bytes.Buffer. Then
// it resets the original response body so it can be used again later on.
func (s *Archiver) readResponseBody(bodyCopy *bytes.Buffer, response *http.Response) error {
	_, err := io.Copy(bodyCopy, response.Body)
	if err != nil {
		return err
	}
	response.Body.Close() // Close the original body
	response.Body = io.NopCloser(bytes.NewReader(bodyCopy.Bytes()))

	return nil
}

// readResponseHeaders reads the response's headers into a bytes.Buffer.
func (s *Archiver) readResponseHeaders(contentBuffer *bytes.Buffer, response *http.Response) {
	contentBuffer.WriteString(
		fmt.Sprintf(
			"HTTP/%d.%d %d %s\r\n",
			response.ProtoMajor,
			response.ProtoMinor,
			response.StatusCode,
			response.Status,
		),
	)

	for key, values := range response.Header {
		for _, value := range values {
			contentBuffer.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
		}
	}
	contentBuffer.WriteString("\r\n")
}

// Close closes the archive and creates the remaining files.
func (s *Archiver) Close() {
	err := s.createIndex()
	if err != nil {
		log.Printf("failed to create index file entry in ZIP: %v", err)
	}

	err = s.createPages()
	if err != nil {
		log.Printf("failed to create pages file entry in ZIP: %v", err)
	}

	s.waczWriter.Close()
	s.file.Close()

	err = s.createDatapackage()
	if err != nil {
		log.Printf("failed to create datapackage.json: %v", err)
		return
	}
}

// Create the index file.
func (s *Archiver) createIndex() error {
	indexWriter, err := s.waczWriter.Create("indexes/index.cdx")
	if err != nil {
		return err
	}

	cdx := []string{}
	for _, entry := range s.cdxjEntries {
		domainParts := strings.Split(entry.url.Hostname(), ".")
		slices.Reverse(domainParts)
		searchableURL := strings.Join(domainParts, ",")
		searchableURL = searchableURL + ")" + entry.url.RequestURI()

		jsonEntry, err := json.Marshal(entry)
		if err != nil {
			log.Printf("failed to json marshal index %v", err)
			continue
		}

		cdxjLine := fmt.Sprintf("%s %s %s\n", searchableURL, entry.time.Format("20060102150405"), jsonEntry)
		cdx = append(cdx, cdxjLine)
	}

	slices.Sort(cdx)
	for _, e := range cdx {
		indexWriter.Write([]byte(e))
	}

	return nil
}

// addPage adds a new page record in the pages.jsonl file.
func (s *Archiver) createPages() error {
	// Add the pages.jsonl and add the file header
	pagesWriter, err := s.waczWriter.Create("pages/pages.jsonl")
	if err != nil {
		return err
	}

	header := `{"format": "json-pages-1.0", "id": "pages", "title": "All Pages"}`
	header += "\n"
	pagesWriter.Write([]byte(header))

	for _, e := range s.cdxjEntries {
		page := PageEntry{
			URL: e.url.String(),
			TS:  e.time.Format(time.RFC3339),
		}

		jsonPage, err := json.Marshal(page)
		if err != nil {
			return err
		}

		jsonPage = append(jsonPage, '\n')
		_, err = pagesWriter.Write(jsonPage)
		if err != nil {
			log.Printf("error adding page %s %v", jsonPage, err)
		}
	}

	return nil
}

// Create the datapackage.json file.
// Opens the zip file and reads all the files to create the resources json along with the hash.
// Then it saves the datapackage and creates the datapackage-digest json file.
func (s *Archiver) createDatapackage() error {
	archive, err := zip.OpenReader(s.file.Name())
	if err != nil {
		log.Printf("Failed to open ZIP archive for reading: %v", err)
		return err
	}
	defer archive.Close()

	datapackage, err := s.getResources(archive)
	if err != nil {
		log.Printf("failed to get zip resources %v", err)
	}

	f, err := os.OpenFile(s.file.Name(), os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	zipWriter := zip.NewWriter(f)
	defer zipWriter.Close()

	// Copy existing files.
	for _, file := range archive.File {
		zipWriter.Copy(file)
	}

	// create datapackage.
	datapackageFile, err := zipWriter.Create("datapackage.json")
	if err != nil {
		return err
	}
	_, err = datapackageFile.Write(datapackage)
	if err != nil {
		return err
	}

	// create datapackage digest.
	datapackageDigest, err := zipWriter.Create("datapackage-digest.json")
	if err != nil {
		return err
	}

	hash := sha256.Sum256(datapackage)
	hashHex := hex.EncodeToString(hash[:])
	digestMap := map[string]string{
		"path": "datapackage.json",
		"hash": "sha256" + hashHex,
	}
	digest, err := json.MarshalIndent(digestMap, "", "  ")
	if err != nil {
		return nil
	}

	_, err = datapackageDigest.Write(digest)
	if err != nil {
		return err
	}

	return nil
}

// getResources returns a []byte with the json data for the datapackage.json file.
func (s *Archiver) getResources(archive *zip.ReadCloser) ([]byte, error) {
	var resources []map[string]interface{}
	for _, file := range archive.File {
		hash, err := s.calculateHash(file)
		if err != nil {
			return []byte{}, err
		}

		resources = append(resources, map[string]interface{}{
			"name":  filepath.Base(file.Name),
			"path":  file.Name,
			"hash":  hash,
			"bytes": file.FileInfo().Size(),
		})
	}

	datapackage := map[string]interface{}{
		"profile":      "data-package",
		"wacz_version": "1.1.1",
		"resources":    resources,
	}

	return json.MarshalIndent(datapackage, "", "  ")
}

// calculateHash returns the hash string of a zip.File.
func (s *Archiver) calculateHash(file *zip.File) (string, error) {
	rc, err := file.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()
	hash := sha256.New()
	_, err = io.Copy(hash, rc)
	if err != nil {
		return "", err
	}

	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), nil
}
