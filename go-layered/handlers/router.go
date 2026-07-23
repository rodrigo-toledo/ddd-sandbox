package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(orderH *OrderHandler, productH *ProductHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Route("/orders", func(r chi.Router) {
		r.Post("/", orderH.PlaceOrder)
		r.Get("/{id}", orderH.GetOrder)
		r.Post("/{id}/ship", orderH.ShipOrder)
		r.Post("/{id}/deliver", orderH.DeliverOrder)
		r.Post("/{id}/return", orderH.RequestReturn)
	})

	r.Route("/products", func(r chi.Router) {
		r.Post("/", productH.CreateProduct)
		r.Get("/{id}", productH.GetProduct)
	})

	return r
}
