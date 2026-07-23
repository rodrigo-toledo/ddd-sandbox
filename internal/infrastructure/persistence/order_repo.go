package persistence

import (
	"context"
	"time"

	"github.com/rodrigotoledo/ddd-sandbox/internal/domain/order"
	"github.com/rodrigotoledo/ddd-sandbox/internal/domain/shared"
	"github.com/rodrigotoledo/ddd-sandbox/internal/infrastructure/persistence/sqlc"
)

type OrderRepo struct {
	q *sqlc.Queries
}

func NewOrderRepo(q *sqlc.Queries) *OrderRepo {
	return &OrderRepo{q: q}
}

func (r *OrderRepo) Save(o *order.Order) error {
	ctx := context.Background()

	existing, err := r.q.GetOrder(ctx, o.ID)
	if err != nil {
		if err := r.q.InsertOrder(ctx, sqlc.InsertOrderParams{
			ID:            o.ID,
			CustomerID:    o.CustomerID,
			Status:        string(o.Status),
			TotalAmount:   o.Total.Amount,
			TotalCurrency: o.Total.Currency,
			PlacedAt:      o.PlacedAt.Format(time.RFC3339),
		}); err != nil {
			return err
		}
		for _, item := range o.Items {
			if err := r.q.InsertOrderItem(ctx, sqlc.InsertOrderItemParams{
				OrderID:           o.ID,
				ProductID:         item.ProductID,
				Quantity:          int64(item.Quantity),
				UnitPriceAmount:   item.UnitPrice.Amount,
				UnitPriceCurrency: item.UnitPrice.Currency,
			}); err != nil {
				return err
			}
		}
		return nil
	}

	_ = existing
	if err := r.q.UpdateOrder(ctx, sqlc.UpdateOrderParams{
		Status:        string(o.Status),
		TotalAmount:   o.Total.Amount,
		TotalCurrency: o.Total.Currency,
		ID:            o.ID,
	}); err != nil {
		return err
	}
	return nil
}

func (r *OrderRepo) FindByID(id string) (*order.Order, error) {
	ctx := context.Background()

	row, err := r.q.GetOrder(ctx, id)
	if err != nil {
		return nil, err
	}

	items, err := r.q.GetOrderItems(ctx, id)
	if err != nil {
		return nil, err
	}

	placedAt, _ := time.Parse(time.RFC3339, row.PlacedAt)
	o := &order.Order{
		ID:         row.ID,
		CustomerID: row.CustomerID,
		Status:     order.Status(row.Status),
		Total:      shared.MustMoney(row.TotalAmount, row.TotalCurrency),
		PlacedAt:   placedAt,
	}
	for _, item := range items {
		o.Items = append(o.Items, order.Item{
			ProductID: item.ProductID,
			Quantity:  int(item.Quantity),
			UnitPrice: shared.MustMoney(item.UnitPriceAmount, item.UnitPriceCurrency),
		})
	}
	return o, nil
}
