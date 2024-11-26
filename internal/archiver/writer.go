package archiver

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
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/slyrz/warc"
)

type Writer struct {
	file         *os.File
	indexEntries []IndexEntry
	waczWriter   *zip.Writer
	warcWriter   *warc.Writer
	warcOffset   int
}

type IndexEntry struct {
	URL          string    `json:"url"`
	Offset       int       `json:"offset"`
	Status       string    `json:"status"`
	Length       int       `json:"length"`
	Mime         string    `json:"mime"`
	Filename     string    `json:"filename"`
	Digest       string    `json:"digest"`
	RecordDigest string    `json:"recordDigest"`
	time         time.Time `json:"-"`
	parsedURL    url.URL   `json:"-"`
}

type PageEntry struct {
	URL string `json:"url"`
	TS  string `json:"ts"`
}

// Returns a new Writer.
// It creates a new wacz file for the given url string.
func NewArchiver(waczPath string) (*Writer, error) {
	waczDir := filepath.Dir(waczPath)

	err := os.MkdirAll(waczDir, 0755)
	if err != nil {
		return nil, err
	}

	// Create the wacz file.
	file, err := os.Create(waczPath)
	if err != nil {
		return nil, err
	}
	waczWriter := zip.NewWriter(file)

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info %w", err)
	}

	header := &zip.FileHeader{
		Name:   "data/data.warc",
		Method: zip.Store, // Store mode, no compression
	}
	header.Modified = fileInfo.ModTime() // Optional: keep original modification time
	// header.UncompressedSize64 = uint64(fileInfo.Size()) // Set uncompressed size

	// Create the warc writer.
	archiveFile, err := waczWriter.CreateHeader(header) //waczWriter.Create("data/data.warc")
	if err != nil {
		log.Printf("failed to create WARC file entry in ZIP: %v", err)
		return nil, err
	}
	warcWriter := warc.NewWriter(archiveFile)

	return &Writer{
		waczWriter: waczWriter,
		file:       file,
		warcWriter: warcWriter,
	}, nil
}

// AddRecord adds a new response record to the warc file and keeps track
// of the added records to create the index once the Writer is closed.
func (s *Writer) AddRecord(response *http.Response) {
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

	recordLen := 0
	wdate := time.Now()
	record := warc.NewRecord()
	record.Header.Set("warc-type", "response")
	record.Header.Set("warc-date", wdate.Format(time.RFC3339))
	record.Header.Set("warc-target-uri", response.Request.URL.String())
	record.Header.Set("content-type", response.Header.Get("Content-Type"))
	record.Header.Set("warc-record-id", fmt.Sprintf("<urn:uuid:%s>", uuid.New().String()))
	record.Content = bytes.NewReader(contentBuffer.Bytes())
	if recordLen, err = s.warcWriter.WriteRecord(record); err != nil {
		log.Printf("failed to write WARC record to archive: %v", err)
		return
	}

	indexEntry := IndexEntry{
		Status:       fmt.Sprintf("%d", response.StatusCode),
		Length:       recordLen,
		Mime:         response.Header.Get("Content-Type"),
		Filename:     "data/data.warc.gz",
		Digest:       fmt.Sprintf("sha-256:%x", sha256.Sum256(bodyCopy.Bytes())),
		RecordDigest: fmt.Sprintf("sha256:%x", sha256.Sum256(contentBuffer.Bytes())),
		Offset:       s.warcOffset,
		time:         wdate,
		parsedURL:    *response.Request.URL,
		URL:          response.Request.URL.String(),
	}
	s.indexEntries = append(s.indexEntries, indexEntry)

	s.warcOffset += recordLen
}

// readResponseBody Reads the http response's body into a bytes.Buffer. Then
// it resets the original response body so it can be used again later on.
func (s *Writer) readResponseBody(bodyCopy *bytes.Buffer, response *http.Response) error {
	_, err := io.Copy(bodyCopy, response.Body)
	if err != nil {
		return err
	}
	response.Body.Close() // Close the original body
	response.Body = io.NopCloser(bytes.NewReader(bodyCopy.Bytes()))

	return nil
}

// readResponseHeaders reads the response's headers into a bytes.Buffer.
func (s *Writer) readResponseHeaders(contentBuffer *bytes.Buffer, response *http.Response) {
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
func (s *Writer) Close() {
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
func (s *Writer) createIndex() error {
	header := &zip.FileHeader{
		Name:   "indexes/index.cdx",
		Method: zip.Store, // Store mode, no compression
	}

	indexWriter, err := s.waczWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	cdx := []string{}
	for _, entry := range s.indexEntries {
		domainParts := strings.Split(entry.parsedURL.Hostname(), ".")
		slices.Reverse(domainParts)
		searchableURL := strings.Join(domainParts, ",")
		searchableURL = searchableURL + ")" + entry.parsedURL.RequestURI()

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
func (s *Writer) createPages() error {
	pagesWriter, err := s.waczWriter.Create("pages/pages.jsonl")
	if err != nil {
		return err
	}

	header := `{"format": "json-pages-1.0", "id": "pages", "title": "All Pages"}`
	header += "\n"
	pagesWriter.Write([]byte(header))

	for _, e := range s.indexEntries {
		page := PageEntry{
			URL: e.parsedURL.String(),
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
func (s *Writer) createDatapackage() error {
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
func (s *Writer) getResources(archive *zip.ReadCloser) ([]byte, error) {
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
func (s *Writer) calculateHash(file *zip.File) (string, error) {
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
