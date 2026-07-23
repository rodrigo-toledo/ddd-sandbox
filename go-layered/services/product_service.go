package services

import (
	"context"
	"errors"

	"github.com/rodrigotoledo/ddd-sandbox/go-layered/models"
	"github.com/rodrigotoledo/ddd-sandbox/go-layered/repositories"
)

var ErrProductNotFound = errors.New("product not found")

type ProductService struct {
	repo *repositories.ProductRepo
}

func NewProductService(repo *repositories.ProductRepo) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) Create(ctx context.Context, id, name string, stock int) (*models.Product, error) {
	p := &models.Product{ID: id, Name: name, Stock: stock}
	if err := s.repo.Save(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *ProductService) GetByID(ctx context.Context, id string) (*models.Product, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrProductNotFound
	}
	return p, nil
}
