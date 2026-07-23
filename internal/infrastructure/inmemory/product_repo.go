package inmemory

import (
	"fmt"

	"github.com/rodrigotoledo/ddd-sandbox/internal/domain/inventory"
)

type ProductRepo struct {
	products map[string]*inventory.Product
}

func NewProductRepo() *ProductRepo {
	return &ProductRepo{products: make(map[string]*inventory.Product)}
}

func (r *ProductRepo) Save(p *inventory.Product) error {
	r.products[p.ID] = p
	return nil
}

func (r *ProductRepo) FindByID(id string) (*inventory.Product, error) {
	p, ok := r.products[id]
	if !ok {
		return nil, fmt.Errorf("product %s not found", id)
	}
	return p, nil
}
