package container_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stjudewashere/seonaut/internal/container"
)

func TestPubSub(t *testing.T) {
	testval := false
	expectedval := true

	broker := container.NewPubSubBroker()

	broker.NewSubscriber("channel", func(m *container.Message) error {
		testval = m.Data.(bool)
		return nil
	})

	broker.Publish("channel", &container.Message{Data: expectedval})

	if testval != expectedval {
		t.Errorf("%v= %v\n", testval, expectedval)
	}
}

func TestPubSubUnsusbscribe(t *testing.T) {
	var counter int
	expected := 1

	broker := container.NewPubSubBroker()

	subscriber := broker.NewSubscriber("channel", func(m *container.Message) error {
		counter++
		return nil
	})

	broker.Publish("channel", &container.Message{})

	broker.Unsubscribe(subscriber)

	broker.Publish("channel", &container.Message{})

	if counter != expected {
		t.Errorf("%v= %v\n", counter, expected)
	}
}

// Test PubSub concurrency.
func TestPubSubConcurrent(t *testing.T) {
	const numSubscribers = 10
	const numMessages = 1000

	broker := container.NewPubSubBroker()

	var receivedMessages int32
	var wg sync.WaitGroup
	for i := 0; i < numSubscribers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			broker.NewSubscriber("channel", func(m *container.Message) error {
				atomic.AddInt32(&receivedMessages, 1)
				return nil
			})
		}(i)
	}
	wg.Wait()

	wg.Add(numMessages)
	for i := 0; i < numMessages; i++ {
		go func(msgID int) {
			defer wg.Done()
			broker.Publish("channel", &container.Message{})
		}(i)
	}
	wg.Wait()

	expectedMessages := numSubscribers * numMessages
	if int(receivedMessages) != expectedMessages {
		t.Errorf("Expected %d messages, but received %d", expectedMessages, receivedMessages)
	}
}
