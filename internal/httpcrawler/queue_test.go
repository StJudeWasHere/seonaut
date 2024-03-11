package httpcrawler_test

import (
	"context"
	"testing"

	"github.com/stjudewashere/seonaut/internal/httpcrawler"
)

func TestFIFO(t *testing.T) {
	el1 := &httpcrawler.RequestMessage{URL: "element 1"}
	el2 := &httpcrawler.RequestMessage{URL: "element 2"}

	queue := httpcrawler.NewQueue(context.Background())
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
}

func TestActiveNotActive(t *testing.T) {
	queue := httpcrawler.NewQueue(context.Background())
	el1 := &httpcrawler.RequestMessage{URL: "element 1"}

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
	queue.Ack(el1.URL)

	active = queue.Active()
	if active != false {
		t.Errorf("Queue should not be active. Is: %v", active)
	}
}
