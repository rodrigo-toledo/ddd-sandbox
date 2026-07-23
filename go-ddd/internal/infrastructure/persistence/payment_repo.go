package persistence

import (
	"context"

	"github.com/rodrigotoledo/ddd-sandbox/go-ddd/internal/domain/payment"
	"github.com/rodrigotoledo/ddd-sandbox/go-ddd/internal/domain/shared"
	"github.com/rodrigotoledo/ddd-sandbox/go-ddd/internal/infrastructure/persistence/sqlc"
)

type PaymentRepo struct {
	q *sqlc.Queries
}

func NewPaymentRepo(q *sqlc.Queries) *PaymentRepo {
	return &PaymentRepo{q: q}
}

func (r *PaymentRepo) Save(p *payment.Payment) error {
	ctx := context.Background()

	_, err := r.q.GetPayment(ctx, p.ID)
	if err != nil {
		return r.q.InsertPayment(ctx, sqlc.InsertPaymentParams{
			ID:             p.ID,
			OrderID:        p.OrderID,
			Amount:         p.Amount.Amount,
			Currency:       p.Amount.Currency,
			Status:         string(p.Status),
			RefundedAmount: p.Refunded.Amount,
		})
	}

	return r.q.UpdatePayment(ctx, sqlc.UpdatePaymentParams{
		Status:         string(p.Status),
		RefundedAmount: p.Refunded.Amount,
		ID:             p.ID,
	})
}

func (r *PaymentRepo) FindByID(id string) (*payment.Payment, error) {
	row, err := r.q.GetPayment(context.Background(), id)
	if err != nil {
		return nil, err
	}
	return mapPayment(row), nil
}

func (r *PaymentRepo) FindByOrderID(orderID string) (*payment.Payment, error) {
	row, err := r.q.GetPaymentByOrderID(context.Background(), orderID)
	if err != nil {
		return nil, err
	}
	return mapPayment(row), nil
}

func mapPayment(row sqlc.Payment) *payment.Payment {
	p := payment.New(row.ID, row.OrderID, shared.MustMoney(row.Amount, row.Currency))
	p.Status = payment.Status(row.Status)
	p.Refunded = shared.MustMoney(row.RefundedAmount, row.Currency)
	return p
}
