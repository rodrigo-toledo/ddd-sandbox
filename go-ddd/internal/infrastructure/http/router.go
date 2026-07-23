package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(h *Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Route("/orders", func(r chi.Router) {
		r.Post("/", h.PlaceOrder)
		r.Get("/{id}", h.GetOrder)
		r.Post("/{id}/ship", h.ShipOrder)
		r.Post("/{id}/deliver", h.DeliverOrder)
		r.Post("/{id}/return", h.RequestReturn)
	})

	r.Route("/products", func(r chi.Router) {
		r.Post("/", h.CreateProduct)
		r.Get("/{id}", h.GetProduct)
	})

	return r
}
