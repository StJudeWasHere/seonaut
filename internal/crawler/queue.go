package crawler

type Queue struct {
	in     chan *RequestMessage
	out    chan *RequestMessage
	ack    chan string
	count  chan int
	active chan bool
	done   chan struct{}
}

func NewQueue() *Queue {
	q := Queue{
		in:     make(chan *RequestMessage),
		out:    make(chan *RequestMessage),
		ack:    make(chan string),
		count:  make(chan int),
		active: make(chan bool),
		done:   make(chan struct{}),
	}

	go q.manage()

	return &q
}

// Manage the queue with push, poll, and acknowledgement of elements in the queue.
func (q *Queue) manage() {
	defer func() {
		close(q.in)
		close(q.out)
		close(q.ack)
		close(q.count)
		close(q.active)
		close(q.done)
	}()

	queue := []*RequestMessage{}
	active := make(map[string]bool)

	var first *RequestMessage
	var out chan *RequestMessage

	for {
		if first == nil && len(queue) > 0 {
			first = queue[0]
			active[first.URL.String()] = true
			queue = queue[1:]
		}

		if first == nil {
			out = nil
		} else {
			out = q.out
		}

		select {
		case <-q.done:
			return
		case q.count <- len(queue):
		case q.active <- (len(active) > 0 || len(queue) > 0):
		case v := <-q.in:
			queue = append(queue, v)
		case out <- first:
			first = nil
		case v := <-q.ack:
			delete(active, v)
		}
	}
}

// Adds a new value to the queue's end.
func (q *Queue) Push(value *RequestMessage) {
	q.in <- value
}

// Returns the first element in the queue.
func (q *Queue) Poll() *RequestMessage {
	return <-q.out
}

// Acknowledges a message has been processed.
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
	return <-q.active
}

// Done stops the queue and closes all of its channels.
func (q *Queue) Done() {
	q.done <- struct{}{}
}
