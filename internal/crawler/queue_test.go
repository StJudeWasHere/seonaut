package crawler_test

import (
	"net/url"
	"testing"

	"github.com/stjudewashere/seonaut/internal/crawler"
)

func TestFIFO(t *testing.T) {
	url1 := new(url.URL)
	url1.Scheme = "https"
	url1.Host = "example.com"
	url1.Path = "element1"

	url2 := new(url.URL)
	url2.Scheme = "https"
	url2.Host = "example.com"
	url2.Path = "element2"

	el1 := &crawler.RequestMessage{URL: url1}
	el2 := &crawler.RequestMessage{URL: url2}

	queue := crawler.NewQueue()
	queue.Push(el1)
	queue.Push(el2)

	p1 := queue.Poll()
	if p1 != el1 {
		t.Errorf("%v != %v", p1, el1)
	}

	p2 := queue.Poll()
	if p2 != el2 {
		t.Errorf("%v != %v", p2, el2)
	}

	queue.Done()
}

func TestActiveNotActive(t *testing.T) {
	queue := crawler.NewQueue()
	url1 := new(url.URL)
	url1.Scheme = "https"
	url1.Host = "example.com"
	url1.Path = "element1"

	el1 := &crawler.RequestMessage{URL: url1}

	queue.Push(el1)

	active := queue.Active()
	if active != true {
		t.Errorf("Queue should be active. Is: %v", active)
	}

	_ = queue.Poll()

	active = queue.Active()
	if active != true {
		t.Errorf("Queue should be active. Is: %v", active)
	}

	// Acknowledge element 1.
	// After the acknowledge, the queue should be empty and not active.
	queue.Ack(el1.URL.String())

	active = queue.Active()
	if active != false {
		t.Errorf("Queue should not be active. Is: %v", active)
	}

	queue.Done()
}
