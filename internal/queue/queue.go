package queue

import (
	"context"
)

type Queue struct {
	in     chan string
	out    chan string
	ack    chan string
	count  chan int
	active chan int
}

func New(ctx context.Context) *Queue {
	q := Queue{
		in:     make(chan string),
		out:    make(chan string),
		ack:    make(chan string),
		count:  make(chan int),
		active: make(chan int),
	}

	go q.manage(ctx)

	return &q
}

// Manage the queue with push, poll, and acknowledgement of elements in the queue.
func (q *Queue) manage(ctx context.Context) {
	defer func() {
		close(q.in)
		close(q.out)
		close(q.ack)
		close(q.count)
		close(q.active)
	}()

	queue := []string{}
	active := make(map[string]bool)

	var first string

	for {
		out := q.out

		if first == "" && len(queue) > 0 {
			first = queue[0]
			active[first] = true
			queue = queue[1:]
		}

		if first == "" {
			out = nil
		}

		select {
		case <-ctx.Done():
			return
		case q.count <- len(queue):
		case q.active <- len(active):
		case v := <-q.in:
			queue = append(queue, v)
		case out <- first:
			first = ""
		case v := <-q.ack:
			delete(active, v)
		}
	}
}

// Adds a new value to the queue's end.
func (q *Queue) Push(value string) {
	q.in <- value
}

// Returns the first element in the queue.
func (q *Queue) Poll() string {
	return <-q.out
}

// Acknwoledges a message has been processed.
func (q *Queue) Ack(s string) {
	q.ack <- s
}

// Returns the number of items currently in the queue.
func (q *Queue) Count() int {
	v, ok := <-q.count
	if !ok {
		return 0
	}

	return v
}

// Active returns true if the queue is not empty or has active elements.
func (q *Queue) Active() bool {
	va := <-q.active
	vc := <-q.count

	return va > 0 || vc > 0
}
