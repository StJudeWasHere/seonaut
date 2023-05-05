package http_crawler

import (
	"context"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

const (
	// Random delay in milliseconds.
	// A random delay up to this value is introduced before new HTTP requests.
	randomDelay = 1500

	// Number of threads a queue will use to crawl a project.
	consumerThreads = 2
)

type HttpCrawler struct {
	urlStream <-chan string
	rStream   chan *ResponseMessage
	client    *Client
}

type ResponseMessage struct {
	URL      string
	Response *http.Response
	Error    error
}

func New(client *Client, urlStream <-chan string) *HttpCrawler {
	return &HttpCrawler{
		urlStream: urlStream,
		rStream:   make(chan *ResponseMessage),
		client:    client,
	}
}

// Crawl starts crawling the URLs received in the urlStream channel and
// sends ResponseMessage of the crawled URLs through the rStream channel.
// It will end when the context is cancelled.
func (c *HttpCrawler) Crawl(ctx context.Context) <-chan *ResponseMessage {
	go func() {
		defer close(c.rStream)

		wg := new(sync.WaitGroup)
		wg.Add(consumerThreads)

		for i := 0; i < consumerThreads; i++ {
			go func() {
				c.consumer(ctx)
				wg.Done()
			}()
		}

		wg.Wait()
	}()

	return c.rStream
}

// Consumer gets URLs from the urlStream until the context is cancelled.
// It adds a random delay between client calls.
func (c *HttpCrawler) consumer(ctx context.Context) {
	for {
		select {
		case u := <-c.urlStream:
			time.Sleep(time.Duration(rand.Intn(randomDelay)) * time.Millisecond)

			rm := &ResponseMessage{URL: u}
			rm.Response, rm.Error = c.client.Get(rm.URL)

			c.rStream <- rm
		case <-ctx.Done():
			return
		}
	}
}
