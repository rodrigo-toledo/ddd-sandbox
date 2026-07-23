package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigotoledo/ddd-sandbox/go-ddd/internal/application/service"
	"github.com/rodrigotoledo/ddd-sandbox/go-ddd/internal/domain/inventory"
	"github.com/rodrigotoledo/ddd-sandbox/go-ddd/internal/domain/order"
	"github.com/rodrigotoledo/ddd-sandbox/go-ddd/internal/domain/shared"
)

type Handler struct {
	svc      *service.OrderService
	orders   order.Repository
	products inventory.Repository
}

func NewHandler(svc *service.OrderService, orders order.Repository, products inventory.Repository) *Handler {
	return &Handler{svc: svc, orders: orders, products: products}
}

type placeOrderRequest struct {
	OrderID    string `json:"order_id"`
	CustomerID string `json:"customer_id"`
	PaymentID  string `json:"payment_id"`
	Items      []struct {
		ProductID string `json:"product_id"`
		Quantity  int    `json:"quantity"`
		UnitPrice int64  `json:"unit_price"`
		Currency  string `json:"currency"`
	} `json:"items"`
}

type orderResponse struct {
	ID         string `json:"id"`
	CustomerID string `json:"customer_id"`
	Status     string `json:"status"`
	Total      int64  `json:"total"`
	Currency   string `json:"currency"`
	Items      []struct {
		ProductID string `json:"product_id"`
		Quantity  int    `json:"quantity"`
		UnitPrice int64  `json:"unit_price"`
	} `json:"items"`
}

type productResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Stock     int    `json:"stock"`
	Available int    `json:"available"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	var req placeOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	items := make([]order.Item, len(req.Items))
	for i, item := range req.Items {
		items[i] = order.Item{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: shared.MustMoney(item.UnitPrice, item.Currency),
		}
	}

	err := h.svc.PlaceOrder(service.PlaceOrderCommand{
		OrderID:    req.OrderID,
		CustomerID: req.CustomerID,
		PaymentID:  req.PaymentID,
		Items:      items,
	})
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, errorResponse{Error: err.Error()})
		return
	}

	o, _ := h.orders.FindByID(req.OrderID)
	writeJSON(w, http.StatusCreated, toOrderResponse(o))
}

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	o, err := h.orders.FindByID(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "order not found"})
		return
	}
	writeJSON(w, http.StatusOK, toOrderResponse(o))
}

func (h *Handler) ShipOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.ShipOrder(id); err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, errorResponse{Error: err.Error()})
		return
	}
	o, _ := h.orders.FindByID(id)
	writeJSON(w, http.StatusOK, toOrderResponse(o))
}

func (h *Handler) DeliverOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.DeliverOrder(id); err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, errorResponse{Error: err.Error()})
		return
	}
	o, _ := h.orders.FindByID(id)
	writeJSON(w, http.StatusOK, toOrderResponse(o))
}

func (h *Handler) RequestReturn(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.RequestReturn(id); err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, errorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "return_requested"})
}

type createProductRequest struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Stock int    `json:"stock"`
}

func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req createProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}
	p := inventory.NewProduct(req.ID, req.Name, req.Stock)
	if err := h.products.Save(p); err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, errorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, productResponse{
		ID:        p.ID,
		Name:      p.Name,
		Stock:     p.Stock,
		Available: p.Available(),
	})
}

func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	p, err := h.products.FindByID(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "product not found"})
		return
	}
	writeJSON(w, http.StatusOK, productResponse{
		ID:        p.ID,
		Name:      p.Name,
		Stock:     p.Stock,
		Available: p.Available(),
	})
}

func toOrderResponse(o *order.Order) orderResponse {
	resp := orderResponse{
		ID:         o.ID,
		CustomerID: o.CustomerID,
		Status:     string(o.Status),
		Total:      o.Total.Amount,
		Currency:   o.Total.Currency,
	}
	for _, item := range o.Items {
		resp.Items = append(resp.Items, struct {
			ProductID string `json:"product_id"`
			Quantity  int    `json:"quantity"`
			UnitPrice int64  `json:"unit_price"`
		}{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice.Amount,
		})
	}
	return resp
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
