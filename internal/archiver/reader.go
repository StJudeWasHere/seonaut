package archiver

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/slyrz/warc"
)

type Reader struct {
	waczPath string
}

func NewReader(waczPath string) *Reader {
	return &Reader{
		waczPath: waczPath,
	}
}

// ReadArchive reads the archive and returns the contents of the warc record for
// the specified URL as a string.
func (s *Reader) ReadArchive(urlStr string) (string, error) {
	wacz, err := zip.OpenReader(s.waczPath)
	if err != nil {
		return "", err
	}
	defer wacz.Close()

	record, err := s.getCDXEntry(wacz, urlStr)
	if err != nil {
		return "", err
	}

	file, err := s.getZipFile(wacz, "data/data.warc")
	if err != nil {
		return "", err
	}

	zipoffset, err := file.DataOffset()
	if err != nil {
		return "", err
	}

	f, err := os.OpenFile(s.waczPath, os.O_RDWR, 0644)
	if err != nil {
		return "", err
	}
	defer f.Close()

	buffer := make([]byte, record.Length)

	_, err = f.ReadAt(buffer, zipoffset+int64(record.Offset))
	if err != nil {
		return "", err
	}

	wr, _ := warc.NewReader(bytes.NewReader(buffer))
	r, err := wr.ReadRecord()
	if err != nil {
		return "", err
	}

	c, _ := io.ReadAll(r.Content)
	return string(c), nil
}

// getCDXEntry Looks for the specified URL in the index file and returns an IndexEntry if found,
// otherwise it returns an error.
func (s *Reader) getCDXEntry(wacz *zip.ReadCloser, urlStr string) (*IndexEntry, error) {
	file, err := s.getZipFile(wacz, "indexes/index.cdx")
	if err != nil {
		return nil, err
	}
	offset, err := file.DataOffset()
	if err != nil {
		return nil, err
	}
	size := file.FileInfo().Size()

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	domainParts := strings.Split(parsedURL.Hostname(), ".")
	slices.Reverse(domainParts)
	searchableURL := strings.Join(domainParts, ",")
	searchableURL = searchableURL + ")" + parsedURL.RequestURI()

	line, err := s.searchFileSegment(offset, size, searchableURL)
	if err != nil {
		return nil, err
	}

	var record IndexEntry

	jsonStart := strings.Index(line, "{")
	if jsonStart == -1 {
		fmt.Println("JSON data not found in line:", line)
		return nil, fmt.Errorf("invalid IndexEntry %s", line)
	}

	// Extract the JSON substring
	jsonData := line[jsonStart:]

	err = json.Unmarshal([]byte(jsonData), &record)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	return &record, nil
}

// getZipFile returns a *zip.File from a wacz file. If not found it returns an error.
func (s *Reader) getZipFile(wacz *zip.ReadCloser, waczFile string) (*zip.File, error) {
	for _, file := range wacz.File {
		if file.Name == waczFile {
			return file, nil
		}
	}

	return nil, errors.New("warc file file not found")
}

// searchFileSegment searches the target string in WACZ file index using binary search.
// It loads the index contents in memory.
func (s *Reader) searchFileSegment(offset, length int64, target string) (string, error) {
	file, err := os.Open(s.waczPath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Seek to the specified offset
	_, err = file.Seek(offset, 0)
	if err != nil {
		return "", fmt.Errorf("failed to seek to offset: %v", err)
	}

	// Read the specified length of bytes
	buffer := make([]byte, length)
	_, err = file.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("failed to read segment: %v", err)
	}

	// Split the buffer into lines making sure to avoid empty lines
	lines := make([]string, 0)
	for _, line := range strings.Split(string(buffer), "\n") {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}

	// Perform binary search on lines in memory using sort.Search
	index := sort.Search(len(lines), func(i int) bool {
		return lines[i] >= target
	})

	// Check if the found line starts with the target prefix
	if index < len(lines) && strings.HasPrefix(lines[index], target) {
		return lines[index], nil // Found the line
	}

	return "", fmt.Errorf("no line starting with '%s' found", target)
}
