package inmemory

import (
	"fmt"

	"github.com/rodrigotoledo/ddd-sandbox/go-ddd/internal/domain/payment"
)

type PaymentRepo struct {
	payments map[string]*payment.Payment
}

func NewPaymentRepo() *PaymentRepo {
	return &PaymentRepo{payments: make(map[string]*payment.Payment)}
}

func (r *PaymentRepo) Save(p *payment.Payment) error {
	r.payments[p.ID] = p
	return nil
}

func (r *PaymentRepo) FindByID(id string) (*payment.Payment, error) {
	p, ok := r.payments[id]
	if !ok {
		return nil, fmt.Errorf("payment %s not found", id)
	}
	return p, nil
}

func (r *PaymentRepo) FindByOrderID(orderID string) (*payment.Payment, error) {
	for _, p := range r.payments {
		if p.OrderID == orderID {
			return p, nil
		}
	}
	return nil, fmt.Errorf("payment for order %s not found", orderID)
}
