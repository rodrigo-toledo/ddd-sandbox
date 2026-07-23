package shared

import "time"

type DomainEvent interface {
	EventName() string
	OccurredOn() time.Time
	AggregateID() string
}

type BaseEvent struct {
	Name        string
	OccurredAt  time.Time
	AggID       string
}

func NewBaseEvent(name, aggregateID string) BaseEvent {
	return BaseEvent{
		Name:       name,
		OccurredAt: time.Now(),
		AggID:      aggregateID,
	}
}

func (e BaseEvent) EventName() string    { return e.Name }
func (e BaseEvent) OccurredOn() time.Time { return e.OccurredAt }
func (e BaseEvent) AggregateID() string   { return e.AggID }
