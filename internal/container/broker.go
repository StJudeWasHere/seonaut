package container

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

func NewPubSubBroker() *Broker {
	return &Broker{
		subscribers: make(map[string][]*Subscriber, 0),
		lock:        &sync.RWMutex{},
	}
}

// Returns a new subsciber to the topic.
func (b *Broker) NewSubscriber(topic string, c func(*Message) error) *Subscriber {
	b.lock.Lock()
	defer b.lock.Unlock()

	s := &Subscriber{
		Id:       uuid.New(),
		Topic:    topic,
		Callback: c,
	}

	b.subscribers[topic] = append(b.subscribers[topic], s)

	return s
}

// Unsubscribes a subscriber.
func (b *Broker) Unsubscribe(s *Subscriber) {
	b.lock.Lock()
	defer b.lock.Unlock()

	subscribers := b.subscribers[s.Topic]

	for i, v := range subscribers {
		if v.Id == s.Id {
			b.subscribers[s.Topic] = append(subscribers[:i], subscribers[i+1:]...)

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
	b.lock.Lock()
	defer b.lock.Unlock()

	subscribers := b.subscribers[topic]

	for i, v := range subscribers {
		err := v.Callback(m)
		if err != nil {
			b.subscribers[topic] = append(subscribers[:i], subscribers[i+1:]...)
		}
	}

	if len(b.subscribers[topic]) == 0 {
		delete(b.subscribers, topic)
	}
}
