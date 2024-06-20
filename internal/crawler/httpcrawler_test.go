package crawler_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stjudewashere/seonaut/internal/crawler"
)

type MockClient struct{}

func (c *MockClient) Get(u string) (*http.Response, error) {
	return nil, nil
}

func (c *MockClient) Head(u string) (*http.Response, error) {
	return nil, nil
}

func (c *MockClient) Do(req *http.Request) (*http.Response, error) {
	return nil, nil
}

func (c *MockClient) GetTTFB(req *http.Response) time.Duration {
	return 0
}

func TestHttpCrawler_Crawl(t *testing.T) {
	mockURLStream := make(chan *crawler.RequestMessage)
	defer close(mockURLStream)

	depth := 5
	testURL := "http://example.com"

	client := &MockClient{}
	httpcrawler := crawler.New(client, mockURLStream)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		mockURLStream <- &crawler.RequestMessage{URL: testURL, Depth: depth}
	}()

	resultStream := httpcrawler.Crawl(ctx)
	result := <-resultStream
	if result.URL != testURL {
		t.Errorf("Expected URL %s, got %s", testURL, result.URL)
	}

	if result.Depth != depth {
		t.Errorf("Expected Depth %d, got %d", depth, result.Depth)
	}

	cancel()
}
