package crawler

type queue struct {
	in    chan interface{}
	out   chan interface{}
	ack   chan interface{}
	count chan interface{}
}

func NewQueue() *queue {
	q := queue{
		in:    make(chan interface{}),
		out:   make(chan interface{}),
		ack:   make(chan interface{}),
		count: make(chan interface{}),
	}
	go q.manage()

	return &q
}

// Manage the push, poll and acknowledege of elements in the queue
func (q *queue) manage() {
	defer func() {
		close(q.in)
		close(q.out)
		close(q.ack)
		close(q.count)
	}()

	queue := []interface{}{}
	active := make(map[string]bool)

	var first interface{}

	for {
		out := q.out

		if first == nil && len(queue) > 0 {
			first = queue[0]
			queue[0] = nil
			queue = queue[1:]
		}

		if first == nil {
			out = nil
		}

		select {
		case q.count <- len(queue):
		case v := <-q.in:
			queue = append(queue, v)
		case out <- first:
			active[first.(string)] = true
			first = nil
		case v := <-q.ack:
			delete(active, v.(string))

			if first == nil && len(queue) == 0 && len(active) == 0 {
				return
			}
		}
	}
}

// Adds a new value to the end of the queue
func (q *queue) Push(value interface{}) {
	q.in <- value
}

// Returns the first element in the queue adn a boolean indicating if operation was ok
func (q *queue) Poll() (interface{}, bool) {
	v, ok := <-q.out

	return v, ok
}

// Acknwoledge a message has been processed
func (q *queue) Ack(s string) {
	q.ack <- s
}

// Returns the number of items currently in the queue
func (q *queue) Count() int {
	v, ok := <-q.count

	if !ok {
		return 0
	}

	return v.(int)
}
