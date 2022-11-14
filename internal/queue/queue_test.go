package queue_test

import (
	"context"
	"testing"

	"github.com/stjudewashere/seonaut/internal/queue"
)

func TestFIFO(t *testing.T) {
	el1 := "element 1"
	el2 := "element 2"

	queue := queue.New(context.Background())
	queue.Push(el1)
	queue.Push(el2)

	p1 := queue.Poll()
	if p1 != el1 {
		t.Errorf("%s != %s", p1, el1)
	}

	p2 := queue.Poll()
	if p2 != el2 {
		t.Errorf("%s != %s", p2, el2)
	}
}

func TestActiveNotActive(t *testing.T) {
	queue := queue.New(context.Background())
	el1 := "element 1"

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
	queue.Ack(el1)

	active = queue.Active()
	if active != false {
		t.Errorf("Queue should not be active. Is: %v", active)
	}
}
