package eventbus

import (
	"testing"

	"github.com/rodrigotoledo/ddd-sandbox/internal/domain/shared"
	"github.com/stretchr/testify/assert"
)

type testEvent struct {
	shared.BaseEvent
	Payload string
}

func TestPublishSubscribe(t *testing.T) {
	bus := NewInMemoryBus()
	var received []shared.DomainEvent

	bus.Subscribe("test.event", func(e shared.DomainEvent) {
		received = append(received, e)
	})

	bus.Publish(testEvent{
		BaseEvent: shared.NewBaseEvent("test.event", "agg-1"),
		Payload:   "hello",
	})

	assert.Len(t, received, 1)
	assert.Equal(t, "test.event", received[0].EventName())
}

func TestMultipleSubscribers(t *testing.T) {
	bus := NewInMemoryBus()
	count := 0

	bus.Subscribe("test.event", func(e shared.DomainEvent) { count++ })
	bus.Subscribe("test.event", func(e shared.DomainEvent) { count++ })

	bus.Publish(testEvent{BaseEvent: shared.NewBaseEvent("test.event", "agg-1")})

	assert.Equal(t, 2, count)
}

func TestNoCrossTalk(t *testing.T) {
	bus := NewInMemoryBus()
	count := 0

	bus.Subscribe("event.a", func(e shared.DomainEvent) { count++ })

	bus.Publish(testEvent{BaseEvent: shared.NewBaseEvent("event.b", "agg-1")})

	assert.Equal(t, 0, count)
}

func TestPublishAll(t *testing.T) {
	bus := NewInMemoryBus()
	count := 0

	bus.Subscribe("test.event", func(e shared.DomainEvent) { count++ })

	events := []shared.DomainEvent{
		testEvent{BaseEvent: shared.NewBaseEvent("test.event", "agg-1")},
		testEvent{BaseEvent: shared.NewBaseEvent("test.event", "agg-2")},
	}
	bus.PublishAll(events)

	assert.Equal(t, 2, count)
}
