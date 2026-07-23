package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/rodrigotoledo/ddd-sandbox/go-layered/db"
	"github.com/rodrigotoledo/ddd-sandbox/go-layered/models"
)

type OrderRepo struct {
	q *db.Queries
}

func NewOrderRepo(database *sql.DB) *OrderRepo {
	return &OrderRepo{q: db.New(database)}
}

func (r *OrderRepo) Save(ctx context.Context, o *models.Order) error {
	_, err := r.q.GetOrder(ctx, o.ID)
	if err != nil {
		if err := r.q.InsertOrder(ctx, db.InsertOrderParams{
			ID:            o.ID,
			CustomerID:    o.CustomerID,
			Status:        string(o.Status),
			TotalAmount:   o.Total,
			TotalCurrency: o.Currency,
			PlacedAt:      time.Now().Format(time.RFC3339),
		}); err != nil {
			return err
		}
		for _, item := range o.Items {
			if err := r.q.InsertOrderItem(ctx, db.InsertOrderItemParams{
				OrderID:           o.ID,
				ProductID:         item.ProductID,
				Quantity:          int64(item.Quantity),
				UnitPriceAmount:   item.UnitPrice,
				UnitPriceCurrency: item.Currency,
			}); err != nil {
				return err
			}
		}
		return nil
	}
	return r.q.UpdateOrderStatus(ctx, db.UpdateOrderStatusParams{
		Status: string(o.Status),
		ID:     o.ID,
	})
}

func (r *OrderRepo) FindByID(ctx context.Context, id string) (*models.Order, error) {
	row, err := r.q.GetOrder(ctx, id)
	if err != nil {
		return nil, err
	}
	o := &models.Order{
		ID:         row.ID,
		CustomerID: row.CustomerID,
		Status:     models.OrderStatus(row.Status),
		Total:      row.TotalAmount,
		Currency:   row.TotalCurrency,
		PlacedAt:   row.PlacedAt,
	}
	items, err := r.q.GetOrderItems(ctx, id)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		o.Items = append(o.Items, models.OrderItem{
			ProductID: item.ProductID,
			Quantity:  int(item.Quantity),
			UnitPrice: item.UnitPriceAmount,
			Currency:  item.UnitPriceCurrency,
		})
	}
	return o, nil
}
