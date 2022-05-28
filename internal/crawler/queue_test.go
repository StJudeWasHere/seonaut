package crawler_test

import (
	"testing"

	"github.com/stjudewashere/seonaut/internal/crawler"
)

func TestFIFO(t *testing.T) {
	el1 := "element 1"
	el2 := "element 2"

	queue := crawler.NewQueue()
	queue.Push(el1)
	queue.Push(el2)

	p1, _ := queue.Poll()
	if p1.(string) != el1 {
		t.Errorf("%s != %s", p1.(string), el1)
	}

	p2, _ := queue.Poll()
	if p2.(string) != el2 {
		t.Errorf("%s != %s", p2.(string), el2)
	}
}

func TestOkNotOk(t *testing.T) {
	queue := crawler.NewQueue()
	el1 := "element 1"

	queue.Push(el1)

	_, ok := queue.Poll()
	if ok != true {
		t.Errorf("%v should be true", ok)
	}

	// Acknowledge element 1.
	// Queue should be now empty and should shutdown itself.
	queue.Ack(el1)

	_, ok = queue.Poll()
	if ok != false {
		t.Errorf("%v should be false", ok)
	}
}
