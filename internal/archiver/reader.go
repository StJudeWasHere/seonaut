package archiver

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
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

func (s *Reader) ReadArchive(urlStr string) (content string) {
	wacz, err := zip.OpenReader(s.waczPath)
	if err != nil {
		log.Printf("failed to open reader %v", err)
		return ""
	}
	defer wacz.Close()

	record, err := s.getCDXEntry(wacz, urlStr)
	if err != nil {
		return ""
	}

	zipoffset, err := s.getWarcOffset(wacz)
	if err != nil {
		return ""
	}

	f, err := os.OpenFile(s.waczPath, os.O_RDWR, 0644)
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()

	buffer := make([]byte, record.Length)

	// Read a specific chunk from the file starting from 'offset'
	_, err = f.ReadAt(buffer, zipoffset+int64(record.Offset))
	if err != nil && err.Error() != "EOF" {
		log.Println(err)
	}
	wr, _ := warc.NewReader(bytes.NewReader(buffer))
	r, err := wr.ReadRecord()
	if err != nil {
		log.Println(err)
	}
	log.Println(string(buffer))

	c, _ := io.ReadAll(r.Content)
	return string(c)
}

func (s *Reader) getCDXEntry(wacz *zip.ReadCloser, urlStr string) (*CDXJEntry, error) {
	indexFile, err := wacz.Open("indexes/index.cdx")
	if err != nil {
		log.Printf("failed to open index file %v", err)
		return nil, err
	}

	var record CDXJEntry
	scanner := bufio.NewScanner(indexFile)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, urlStr) {
			// Find the JSON part of the line by locating the first '{' character
			jsonStart := strings.Index(line, "{")
			if jsonStart == -1 {
				fmt.Println("JSON data not found in line:", line)
				continue
			}

			// Extract the JSON substring
			jsonData := line[jsonStart:]

			// Parse the JSON data into the Record struct
			err := json.Unmarshal([]byte(jsonData), &record)
			if err != nil {
				fmt.Println("Error parsing JSON:", err)
				continue
			}

			return &record, nil
		}
	}

	return nil, errors.New("URL not found in index file")
}

func (s *Reader) getWarcOffset(wacz *zip.ReadCloser) (int64, error) {
	var zipOffset int64
	var err error
	for _, file := range wacz.File {
		if file.Name == "data/data.warc" {
			zipOffset, err = file.DataOffset()
			if err != nil {
				return zipOffset, err
			}
			return zipOffset, nil
		}
	}

	return zipOffset, errors.New("warc file file not found")
}
