package persistence

import (
	"context"

	"github.com/rodrigotoledo/ddd-sandbox/internal/domain/inventory"
	"github.com/rodrigotoledo/ddd-sandbox/internal/infrastructure/persistence/sqlc"
)

type ProductRepo struct {
	q *sqlc.Queries
}

func NewProductRepo(q *sqlc.Queries) *ProductRepo {
	return &ProductRepo{q: q}
}

func (r *ProductRepo) Save(p *inventory.Product) error {
	ctx := context.Background()

	_, err := r.q.GetProduct(ctx, p.ID)
	if err != nil {
		if err := r.q.InsertProduct(ctx, sqlc.InsertProductParams{
			ID:    p.ID,
			Name:  p.Name,
			Stock: int64(p.Stock),
		}); err != nil {
			return err
		}
	} else {
		if err := r.q.UpdateProduct(ctx, sqlc.UpdateProductParams{
			Name:  p.Name,
			Stock: int64(p.Stock),
			ID:    p.ID,
		}); err != nil {
			return err
		}
	}

	rows, err := r.q.GetReservations(ctx, p.ID)
	if err != nil {
		return err
	}
	existing := make(map[string]bool)
	for _, row := range rows {
		existing[row.OrderID] = true
	}

	current := make(map[string]bool)
	for _, res := range p.Reservations {
		current[res.OrderID] = true
		if !existing[res.OrderID] {
			if err := r.q.InsertReservation(ctx, sqlc.InsertReservationParams{
				ProductID: p.ID,
				OrderID:   res.OrderID,
				Quantity:  int64(res.Quantity),
			}); err != nil {
				return err
			}
		}
	}

	for orderID := range existing {
		if !current[orderID] {
			if err := r.q.DeleteReservation(ctx, sqlc.DeleteReservationParams{
				ProductID: p.ID,
				OrderID:   orderID,
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *ProductRepo) FindByID(id string) (*inventory.Product, error) {
	ctx := context.Background()

	row, err := r.q.GetProduct(ctx, id)
	if err != nil {
		return nil, err
	}

	p := inventory.NewProduct(row.ID, row.Name, int(row.Stock))

	reservations, err := r.q.GetReservations(ctx, id)
	if err != nil {
		return nil, err
	}
	for _, res := range reservations {
		p.Reservations = append(p.Reservations, inventory.Reservation{
			OrderID:  res.OrderID,
			Quantity: int(res.Quantity),
		})
	}

	return p, nil
}
