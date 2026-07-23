package order

import (
	"errors"
	"time"

	"github.com/rodrigotoledo/ddd-sandbox/go-ddd/internal/domain/shared"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusConfirmed Status = "confirmed"
	StatusShipped   Status = "shipped"
	StatusDelivered Status = "delivered"
	StatusCancelled Status = "cancelled"
)

type Item struct {
	ProductID string
	Quantity  int
	UnitPrice shared.Money
}

type Order struct {
	ID         string
	CustomerID string
	Items      []Item
	Status     Status
	Total      shared.Money
	PlacedAt   time.Time
	events     []shared.DomainEvent
}

var (
	ErrEmptyOrder       = errors.New("order must have at least one item")
	ErrAlreadyConfirmed = errors.New("order already confirmed")
	ErrNotConfirmed     = errors.New("order not confirmed")
	ErrNotShipped       = errors.New("order not shipped")
	ErrAlreadyCancelled = errors.New("order already cancelled")
	ErrCannotCancel     = errors.New("cannot cancel order in current state")
)

func New(id, customerID string, items []Item) (*Order, error) {
	if len(items) == 0 {
		return nil, ErrEmptyOrder
	}

	total := shared.MustMoney(0, items[0].UnitPrice.Currency)
	for _, item := range items {
		lineTotal := item.UnitPrice.Multiply(item.Quantity)
		var err error
		total, err = total.Add(lineTotal)
		if err != nil {
			return nil, err
		}
	}

	o := &Order{
		ID:         id,
		CustomerID: customerID,
		Items:      items,
		Status:     StatusPending,
		Total:      total,
		PlacedAt:   time.Now(),
	}
	o.addEvent(OrderPlaced{
		BaseEvent:  shared.NewBaseEvent("order.placed", id),
		CustomerID: customerID,
		Total:      total,
		Items:      items,
	})
	return o, nil
}

func (o *Order) Confirm() error {
	if o.Status != StatusPending {
		return ErrAlreadyConfirmed
	}
	o.Status = StatusConfirmed
	o.addEvent(OrderConfirmed{
		BaseEvent: shared.NewBaseEvent("order.confirmed", o.ID),
		Total:     o.Total,
	})
	return nil
}

func (o *Order) Ship() error {
	if o.Status != StatusConfirmed {
		return ErrNotConfirmed
	}
	o.Status = StatusShipped
	o.addEvent(OrderShipped{
		BaseEvent: shared.NewBaseEvent("order.shipped", o.ID),
	})
	return nil
}

func (o *Order) Deliver() error {
	if o.Status != StatusShipped {
		return ErrNotShipped
	}
	o.Status = StatusDelivered
	o.addEvent(OrderDelivered{
		BaseEvent: shared.NewBaseEvent("order.delivered", o.ID),
	})
	return nil
}

func (o *Order) Cancel() error {
	if o.Status == StatusCancelled {
		return ErrAlreadyCancelled
	}
	if o.Status != StatusPending {
		return ErrCannotCancel
	}
	o.Status = StatusCancelled
	o.addEvent(OrderCancelled{
		BaseEvent: shared.NewBaseEvent("order.cancelled", o.ID),
	})
	return nil
}

func (o *Order) Events() []shared.DomainEvent {
	return o.events
}

func (o *Order) ClearEvents() {
	o.events = nil
}

func (o *Order) addEvent(e shared.DomainEvent) {
	o.events = append(o.events, e)
}
