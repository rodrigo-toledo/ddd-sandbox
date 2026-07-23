package inmemory

import (
	"fmt"

	"github.com/rodrigotoledo/ddd-sandbox/go-ddd/internal/domain/order"
)

type OrderRepo struct {
	orders map[string]*order.Order
}

func NewOrderRepo() *OrderRepo {
	return &OrderRepo{orders: make(map[string]*order.Order)}
}

func (r *OrderRepo) Save(o *order.Order) error {
	r.orders[o.ID] = o
	return nil
}

func (r *OrderRepo) FindByID(id string) (*order.Order, error) {
	o, ok := r.orders[id]
	if !ok {
		return nil, fmt.Errorf("order %s not found", id)
	}
	return o, nil
}
