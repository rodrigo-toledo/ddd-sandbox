package repositories

import (
	"context"
	"database/sql"

	"github.com/rodrigotoledo/ddd-sandbox/go-layered/db"
	"github.com/rodrigotoledo/ddd-sandbox/go-layered/models"
)

type ProductRepo struct {
	q *db.Queries
}

func NewProductRepo(database *sql.DB) *ProductRepo {
	return &ProductRepo{q: db.New(database)}
}

func (r *ProductRepo) Save(ctx context.Context, p *models.Product) error {
	_, err := r.q.GetProduct(ctx, p.ID)
	if err != nil {
		return r.q.InsertProduct(ctx, db.InsertProductParams{
			ID:    p.ID,
			Name:  p.Name,
			Stock: int64(p.Stock),
		})
	}
	return r.q.UpdateProductStock(ctx, db.UpdateProductStockParams{
		Stock: int64(p.Stock),
		ID:    p.ID,
	})
}

func (r *ProductRepo) FindByID(ctx context.Context, id string) (*models.Product, error) {
	row, err := r.q.GetProduct(ctx, id)
	if err != nil {
		return nil, err
	}
	return &models.Product{
		ID:    row.ID,
		Name:  row.Name,
		Stock: int(row.Stock),
	}, nil
}

func (r *ProductRepo) Reserve(ctx context.Context, productID, orderID string, quantity int) error {
	return r.q.InsertReservation(ctx, db.InsertReservationParams{
		ProductID: productID,
		OrderID:   orderID,
		Quantity:  int64(quantity),
	})
}

func (r *ProductRepo) Release(ctx context.Context, productID, orderID string) error {
	return r.q.DeleteReservation(ctx, db.DeleteReservationParams{
		ProductID: productID,
		OrderID:   orderID,
	})
}

func (r *ProductRepo) GetReservations(ctx context.Context, productID string) ([]models.Reservation, error) {
	rows, err := r.q.GetReservations(ctx, productID)
	if err != nil {
		return nil, err
	}
	var reservations []models.Reservation
	for _, row := range rows {
		reservations = append(reservations, models.Reservation{
			ProductID: row.ProductID,
			OrderID:   row.OrderID,
			Quantity:  int(row.Quantity),
		})
	}
	return reservations, nil
}
