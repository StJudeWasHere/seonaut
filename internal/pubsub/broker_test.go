package pubsub_test

import (
	"testing"

	"github.com/stjudewashere/seonaut/internal/pubsub"
)

func TestPubSub(t *testing.T) {
	testval := false
	expectedval := true

	broker := pubsub.NewBroker()

	broker.NewSubscriber("channel", func(m *pubsub.Message) error {
		testval = m.Data.(bool)
		return nil
	})

	broker.Publish("channel", &pubsub.Message{Data: expectedval})

	if testval != expectedval {
		t.Errorf("%v= %v\n", testval, expectedval)
	}
}

func TestPubSubUnsusbscribe(t *testing.T) {
	var counter int
	expected := 1

	broker := pubsub.NewBroker()

	subscriber := broker.NewSubscriber("channel", func(m *pubsub.Message) error {
		counter++
		return nil
	})

	broker.Publish("channel", &pubsub.Message{})

	broker.Unsubscribe(subscriber)

	broker.Publish("channel", &pubsub.Message{})

	if counter != expected {
		t.Errorf("%v= %v\n", counter, expected)
	}
}
