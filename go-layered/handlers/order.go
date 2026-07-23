package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigotoledo/ddd-sandbox/go-layered/models"
	"github.com/rodrigotoledo/ddd-sandbox/go-layered/services"
)

type OrderHandler struct {
	svc *services.OrderService
}

func NewOrderHandler(svc *services.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
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

type errorResponse struct {
	Error string `json:"error"`
}

func (h *OrderHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	var req placeOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	items := make([]models.OrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = models.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
			Currency:  item.Currency,
		}
	}

	o, err := h.svc.PlaceOrder(r.Context(), services.PlaceOrderInput{
		OrderID:    req.OrderID,
		CustomerID: req.CustomerID,
		PaymentID:  req.PaymentID,
		Items:      items,
	})
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, errorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, toOrderResponse(o))
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	o, err := h.svc.GetOrder(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "order not found"})
		return
	}
	writeJSON(w, http.StatusOK, toOrderResponse(o))
}

func (h *OrderHandler) ShipOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	o, err := h.svc.ShipOrder(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, errorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, toOrderResponse(o))
}

func (h *OrderHandler) DeliverOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	o, err := h.svc.DeliverOrder(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, errorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, toOrderResponse(o))
}

func (h *OrderHandler) RequestReturn(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.RequestReturn(r.Context(), id, time.Now()); err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, errorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "return_requested"})
}

func toOrderResponse(o *models.Order) orderResponse {
	resp := orderResponse{
		ID:         o.ID,
		CustomerID: o.CustomerID,
		Status:     string(o.Status),
		Total:      o.Total,
		Currency:   o.Currency,
	}
	for _, item := range o.Items {
		resp.Items = append(resp.Items, struct {
			ProductID string `json:"product_id"`
			Quantity  int    `json:"quantity"`
			UnitPrice int64  `json:"unit_price"`
		}{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		})
	}
	return resp
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
