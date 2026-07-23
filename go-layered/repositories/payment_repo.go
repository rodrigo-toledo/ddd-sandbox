package repositories

import (
	"context"
	"database/sql"

	"github.com/rodrigotoledo/ddd-sandbox/go-layered/db"
	"github.com/rodrigotoledo/ddd-sandbox/go-layered/models"
)

type PaymentRepo struct {
	q *db.Queries
}

func NewPaymentRepo(database *sql.DB) *PaymentRepo {
	return &PaymentRepo{q: db.New(database)}
}

func (r *PaymentRepo) Save(ctx context.Context, p *models.Payment) error {
	_, err := r.q.GetPayment(ctx, p.ID)
	if err != nil {
		return r.q.InsertPayment(ctx, db.InsertPaymentParams{
			ID:             p.ID,
			OrderID:        p.OrderID,
			Amount:         p.Amount,
			Currency:       p.Currency,
			Status:         string(p.Status),
			RefundedAmount: p.Refunded,
		})
	}
	return r.q.UpdatePaymentStatus(ctx, db.UpdatePaymentStatusParams{
		Status:         string(p.Status),
		RefundedAmount: p.Refunded,
		ID:             p.ID,
	})
}

func (r *PaymentRepo) FindByID(ctx context.Context, id string) (*models.Payment, error) {
	row, err := r.q.GetPayment(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapPayment(row), nil
}

func (r *PaymentRepo) FindByOrderID(ctx context.Context, orderID string) (*models.Payment, error) {
	row, err := r.q.GetPaymentByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	return mapPayment(row), nil
}

func mapPayment(row db.Payment) *models.Payment {
	return &models.Payment{
		ID:       row.ID,
		OrderID:  row.OrderID,
		Amount:   row.Amount,
		Currency: row.Currency,
		Status:   models.PaymentStatus(row.Status),
		Refunded: row.RefundedAmount,
	}
}
