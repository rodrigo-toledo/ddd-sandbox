package payment

import (
	"errors"

	"github.com/rodrigotoledo/ddd-sandbox/internal/domain/shared"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusAuthorized Status = "authorized"
	StatusCaptured   Status = "captured"
	StatusVoided     Status = "voided"
	StatusRefunded   Status = "refunded"
)

type Payment struct {
	ID         string
	OrderID    string
	Amount     shared.Money
	Status     Status
	Refunded   shared.Money
	events     []shared.DomainEvent
}

var (
	ErrNotAuthorized   = errors.New("payment not authorized")
	ErrAlreadyCaptured = errors.New("payment already captured")
	ErrAlreadyVoided   = errors.New("payment already voided")
	ErrRefundExceeds   = errors.New("refund exceeds captured amount")
	ErrNotCaptured     = errors.New("payment not captured")
)

func New(id, orderID string, amount shared.Money) *Payment {
	return &Payment{
		ID:       id,
		OrderID:  orderID,
		Amount:   amount,
		Status:   StatusPending,
		Refunded: shared.MustMoney(0, amount.Currency),
	}
}

func (p *Payment) Authorize() error {
	if p.Status != StatusPending {
		return ErrAlreadyVoided
	}
	p.Status = StatusAuthorized
	p.addEvent(PaymentAuthorized{
		BaseEvent: shared.NewBaseEvent("payment.authorized", p.ID),
		OrderID:   p.OrderID,
		Amount:    p.Amount,
	})
	return nil
}

func (p *Payment) Capture() error {
	if p.Status != StatusAuthorized {
		return ErrNotAuthorized
	}
	p.Status = StatusCaptured
	p.addEvent(PaymentCaptured{
		BaseEvent: shared.NewBaseEvent("payment.captured", p.ID),
		OrderID:   p.OrderID,
		Amount:    p.Amount,
	})
	return nil
}

func (p *Payment) Void() error {
	if p.Status == StatusCaptured {
		return ErrAlreadyCaptured
	}
	if p.Status == StatusVoided {
		return ErrAlreadyVoided
	}
	if p.Status != StatusAuthorized {
		return ErrNotAuthorized
	}
	p.Status = StatusVoided
	p.addEvent(PaymentVoided{
		BaseEvent: shared.NewBaseEvent("payment.voided", p.ID),
		OrderID:   p.OrderID,
		Amount:    p.Amount,
	})
	return nil
}

func (p *Payment) Refund(amount shared.Money) error {
	if p.Status != StatusCaptured {
		return ErrNotCaptured
	}
	newRefunded, err := p.Refunded.Add(amount)
	if err != nil {
		return err
	}
	if newRefunded.Amount > p.Amount.Amount {
		return ErrRefundExceeds
	}
	p.Refunded = newRefunded
	if p.Refunded.Amount == p.Amount.Amount {
		p.Status = StatusRefunded
	}
	p.addEvent(PaymentRefunded{
		BaseEvent: shared.NewBaseEvent("payment.refunded", p.ID),
		OrderID:   p.OrderID,
		Amount:    amount,
	})
	return nil
}

func (p *Payment) Events() []shared.DomainEvent {
	return p.events
}

func (p *Payment) ClearEvents() {
	p.events = nil
}

func (p *Payment) addEvent(e shared.DomainEvent) {
	p.events = append(p.events, e)
}
