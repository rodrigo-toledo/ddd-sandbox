package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rodrigotoledo/ddd-sandbox/go-layered/models"
	"github.com/rodrigotoledo/ddd-sandbox/go-layered/repositories"
)

var (
	ErrEmptyOrder       = errors.New("order must have at least one item")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrOrderNotFound    = errors.New("order not found")
	ErrInvalidTransition = errors.New("invalid status transition")
	ErrReturnWindowExpired = errors.New("return window expired")
)

type PaymentGateway interface {
	Authorize(p *models.Payment) error
	Capture(p *models.Payment) error
	Void(p *models.Payment) error
}

type OrderService struct {
	orders   *repositories.OrderRepo
	products *repositories.ProductRepo
	payments *repositories.PaymentRepo
	gateway  PaymentGateway
}

func NewOrderService(
	orders *repositories.OrderRepo,
	products *repositories.ProductRepo,
	payments *repositories.PaymentRepo,
	gateway PaymentGateway,
) *OrderService {
	return &OrderService{
		orders:   orders,
		products: products,
		payments: payments,
		gateway:  gateway,
	}
}

type PlaceOrderInput struct {
	OrderID    string
	CustomerID string
	PaymentID  string
	Items      []models.OrderItem
}

func (s *OrderService) PlaceOrder(ctx context.Context, input PlaceOrderInput) (*models.Order, error) {
	if len(input.Items) == 0 {
		return nil, ErrEmptyOrder
	}

	var total int64
	currency := input.Items[0].Currency
	for _, item := range input.Items {
		total += item.UnitPrice * int64(item.Quantity)
	}

	o := &models.Order{
		ID:         input.OrderID,
		CustomerID: input.CustomerID,
		Status:     models.OrderPending,
		Total:      total,
		Currency:   currency,
		PlacedAt:   time.Now().Format(time.RFC3339),
		Items:      input.Items,
	}
	if err := s.orders.Save(ctx, o); err != nil {
		return nil, err
	}

	var reservedProducts []string
	for _, item := range input.Items {
		p, err := s.products.FindByID(ctx, item.ProductID)
		if err != nil {
			s.compensate(ctx, o, reservedProducts, "")
			return nil, fmt.Errorf("product %s: %w", item.ProductID, err)
		}
		reservations, _ := s.products.GetReservations(ctx, item.ProductID)
		reserved := 0
		for _, r := range reservations {
			reserved += r.Quantity
		}
		if p.Stock-reserved < item.Quantity {
			s.compensate(ctx, o, reservedProducts, "")
			return nil, ErrInsufficientStock
		}
		if err := s.products.Reserve(ctx, item.ProductID, input.OrderID, item.Quantity); err != nil {
			s.compensate(ctx, o, reservedProducts, "")
			return nil, err
		}
		reservedProducts = append(reservedProducts, item.ProductID)
	}

	payment := &models.Payment{
		ID:       input.PaymentID,
		OrderID:  input.OrderID,
		Amount:   total,
		Currency: currency,
		Status:   models.PaymentPending,
	}
	if err := s.payments.Save(ctx, payment); err != nil {
		s.compensate(ctx, o, reservedProducts, "")
		return nil, err
	}

	if err := s.gateway.Authorize(payment); err != nil {
		s.compensate(ctx, o, reservedProducts, payment.ID)
		return nil, fmt.Errorf("authorize payment: %w", err)
	}
	if err := s.payments.Save(ctx, payment); err != nil {
		s.compensate(ctx, o, reservedProducts, payment.ID)
		return nil, err
	}

	if err := s.gateway.Capture(payment); err != nil {
		s.compensate(ctx, o, reservedProducts, payment.ID)
		return nil, fmt.Errorf("capture payment: %w", err)
	}
	if err := s.payments.Save(ctx, payment); err != nil {
		s.compensate(ctx, o, reservedProducts, payment.ID)
		return nil, err
	}

	for _, item := range input.Items {
		p, _ := s.products.FindByID(ctx, item.ProductID)
		p.Stock -= item.Quantity
		s.products.Save(ctx, p)
		s.products.Release(ctx, item.ProductID, input.OrderID)
	}

	o.Status = models.OrderConfirmed
	if err := s.orders.Save(ctx, o); err != nil {
		return nil, err
	}

	return o, nil
}

func (s *OrderService) compensate(ctx context.Context, o *models.Order, reservedProducts []string, paymentID string) {
	if paymentID != "" {
		p, err := s.payments.FindByID(ctx, paymentID)
		if err == nil && p.Status == models.PaymentAuthorized {
			s.gateway.Void(p)
			s.payments.Save(ctx, p)
		}
	}

	for _, productID := range reservedProducts {
		s.products.Release(ctx, productID, o.ID)
	}

	o.Status = models.OrderCancelled
	s.orders.Save(ctx, o)
}

func (s *OrderService) GetOrder(ctx context.Context, id string) (*models.Order, error) {
	o, err := s.orders.FindByID(ctx, id)
	if err != nil {
		return nil, ErrOrderNotFound
	}
	return o, nil
}

func (s *OrderService) ShipOrder(ctx context.Context, id string) (*models.Order, error) {
	o, err := s.orders.FindByID(ctx, id)
	if err != nil {
		return nil, ErrOrderNotFound
	}
	if o.Status != models.OrderConfirmed {
		return nil, fmt.Errorf("%w: cannot ship order in %s status", ErrInvalidTransition, o.Status)
	}
	o.Status = models.OrderShipped
	if err := s.orders.Save(ctx, o); err != nil {
		return nil, err
	}
	return o, nil
}

func (s *OrderService) DeliverOrder(ctx context.Context, id string) (*models.Order, error) {
	o, err := s.orders.FindByID(ctx, id)
	if err != nil {
		return nil, ErrOrderNotFound
	}
	if o.Status != models.OrderShipped {
		return nil, fmt.Errorf("%w: cannot deliver order in %s status", ErrInvalidTransition, o.Status)
	}
	o.Status = models.OrderDelivered
	if err := s.orders.Save(ctx, o); err != nil {
		return nil, err
	}
	return o, nil
}

func (s *OrderService) RequestReturn(ctx context.Context, id string, now time.Time) error {
	o, err := s.orders.FindByID(ctx, id)
	if err != nil {
		return ErrOrderNotFound
	}
	if o.Status != models.OrderDelivered {
		return fmt.Errorf("%w: cannot return order in %s status", ErrInvalidTransition, o.Status)
	}

	payment, err := s.payments.FindByOrderID(ctx, id)
	if err != nil {
		return err
	}

	deliveredAt, _ := time.Parse(time.RFC3339, o.PlacedAt)
	if now.After(deliveredAt.Add(30 * 24 * time.Hour)) {
		return ErrReturnWindowExpired
	}

	payment.Status = models.PaymentRefunded
	payment.Refunded = payment.Amount
	return s.payments.Save(ctx, payment)
}
