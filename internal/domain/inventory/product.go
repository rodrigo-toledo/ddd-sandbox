package inventory

import (
	"errors"

	"github.com/rodrigotoledo/ddd-sandbox/internal/domain/shared"
)

type Reservation struct {
	OrderID  string
	Quantity int
}

type Product struct {
	ID           string
	Name         string
	Stock        int
	Reservations []Reservation
	events       []shared.DomainEvent
}

var (
	ErrInsufficientStock  = errors.New("insufficient stock")
	ErrReservationNotFound = errors.New("reservation not found")
	ErrAlreadyReserved    = errors.New("order already has a reservation")
)

func NewProduct(id, name string, stock int) *Product {
	return &Product{
		ID:    id,
		Name:  name,
		Stock: stock,
	}
}

func (p *Product) Available() int {
	reserved := 0
	for _, r := range p.Reservations {
		reserved += r.Quantity
	}
	return p.Stock - reserved
}

func (p *Product) Reserve(orderID string, quantity int) error {
	for _, r := range p.Reservations {
		if r.OrderID == orderID {
			return ErrAlreadyReserved
		}
	}
	if p.Available() < quantity {
		return ErrInsufficientStock
	}
	p.Reservations = append(p.Reservations, Reservation{OrderID: orderID, Quantity: quantity})
	p.addEvent(InventoryReserved{
		BaseEvent: shared.NewBaseEvent("inventory.reserved", p.ID),
		OrderID:   orderID,
		Quantity:  quantity,
	})
	return nil
}

func (p *Product) Release(orderID string) error {
	idx := -1
	for i, r := range p.Reservations {
		if r.OrderID == orderID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return ErrReservationNotFound
	}
	quantity := p.Reservations[idx].Quantity
	p.Reservations = append(p.Reservations[:idx], p.Reservations[idx+1:]...)
	p.addEvent(InventoryReleased{
		BaseEvent: shared.NewBaseEvent("inventory.released", p.ID),
		OrderID:   orderID,
		Quantity:  quantity,
	})
	return nil
}

func (p *Product) ConfirmReservation(orderID string) error {
	idx := -1
	for i, r := range p.Reservations {
		if r.OrderID == orderID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return ErrReservationNotFound
	}
	quantity := p.Reservations[idx].Quantity
	p.Stock -= quantity
	p.Reservations = append(p.Reservations[:idx], p.Reservations[idx+1:]...)
	p.addEvent(ReservationConfirmed{
		BaseEvent: shared.NewBaseEvent("inventory.confirmed", p.ID),
		OrderID:   orderID,
		Quantity:  quantity,
	})
	return nil
}

func (p *Product) Events() []shared.DomainEvent {
	return p.events
}

func (p *Product) ClearEvents() {
	p.events = nil
}

func (p *Product) addEvent(e shared.DomainEvent) {
	p.events = append(p.events, e)
}
