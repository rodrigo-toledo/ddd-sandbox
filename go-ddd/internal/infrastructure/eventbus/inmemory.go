package eventbus

import (
	"sync"

	"github.com/rodrigotoledo/ddd-sandbox/go-ddd/internal/domain/shared"
)

type Handler func(event shared.DomainEvent)

type EventBus interface {
	Subscribe(eventName string, handler Handler)
	Publish(event shared.DomainEvent)
	PublishAll(events []shared.DomainEvent)
}

type InMemoryBus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

func NewInMemoryBus() *InMemoryBus {
	return &InMemoryBus{
		handlers: make(map[string][]Handler),
	}
}

func (b *InMemoryBus) Subscribe(eventName string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventName] = append(b.handlers[eventName], handler)
}

func (b *InMemoryBus) Publish(event shared.DomainEvent) {
	b.mu.RLock()
	handlers := b.handlers[event.EventName()]
	b.mu.RUnlock()

	for _, h := range handlers {
		h(event)
	}
}

func (b *InMemoryBus) PublishAll(events []shared.DomainEvent) {
	for _, e := range events {
		b.Publish(e)
	}
}
