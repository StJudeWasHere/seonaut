package pubsub

import (
	"sync"

	"github.com/google/uuid"
)

type Subscriber struct {
	Id       uuid.UUID
	Topic    string
	Callback func(*Message) error
}

type Broker struct {
	subscribers map[string][]*Subscriber
	lock        *sync.RWMutex
}

type Message struct {
	Name string
	Data interface{}
}

func New() *Broker {
	return &Broker{
		subscribers: make(map[string][]*Subscriber, 0),
		lock:        &sync.RWMutex{},
	}
}

// Returns a new subsciber to the topic.
func (b *Broker) NewSubscriber(topic string, c func(*Message) error) *Subscriber {
	s := &Subscriber{
		Id:       uuid.New(),
		Topic:    topic,
		Callback: c,
	}

	b.lock.Lock()
	b.subscribers[topic] = append(b.subscribers[topic], s)
	b.lock.Unlock()

	return s
}

// Unsubscribes a subscriber.
func (b *Broker) Unsubscribe(s *Subscriber) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	subscribers := b.subscribers[s.Topic]

	for i, v := range subscribers {
		if v.Id == s.Id {
			subs := b.subscribers[s.Topic]
			r := make([]*Subscriber, 0)
			r = append(r, subs[:i]...)
			r = append(r, subs[i+1:]...)

			b.subscribers[s.Topic] = r

			// The topic is removed once there are no more subscribers.
			if len(b.subscribers[s.Topic]) == 0 {
				delete(b.subscribers, s.Topic)
			}

			break
		}
	}
}

// Publishes a message to all subscribers of a topic.
func (b *Broker) Publish(topic string, m *Message) {
	b.lock.RLock()
	subscribers := b.subscribers[topic]
	b.lock.RUnlock()

	for _, v := range subscribers {
		err := v.Callback(m)
		if err != nil {
			b.Unsubscribe(v)
		}
	}
}
