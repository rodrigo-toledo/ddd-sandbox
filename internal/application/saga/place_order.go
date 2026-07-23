package saga

import (
	"fmt"

	"github.com/rodrigotoledo/ddd-sandbox/internal/domain/inventory"
	"github.com/rodrigotoledo/ddd-sandbox/internal/domain/order"
	"github.com/rodrigotoledo/ddd-sandbox/internal/domain/payment"
	"github.com/rodrigotoledo/ddd-sandbox/internal/domain/shared"
	"github.com/rodrigotoledo/ddd-sandbox/internal/infrastructure/eventbus"
)

type PaymentGateway interface {
	Authorize(p *payment.Payment) error
	Capture(p *payment.Payment) error
	Void(p *payment.Payment) error
}

type PlaceOrderSaga struct {
	orders   order.Repository
	products inventory.Repository
	payments payment.Repository
	gateway  PaymentGateway
	bus      eventbus.EventBus
}

func NewPlaceOrderSaga(
	orders order.Repository,
	products inventory.Repository,
	payments payment.Repository,
	gateway PaymentGateway,
	bus eventbus.EventBus,
) *PlaceOrderSaga {
	return &PlaceOrderSaga{
		orders:   orders,
		products: products,
		payments: payments,
		gateway:  gateway,
		bus:      bus,
	}
}

type sagaState struct {
	orderID          string
	reservedProducts []string
	paymentID        string
}

func (s *PlaceOrderSaga) Execute(orderID, paymentID string) error {
	state := &sagaState{orderID: orderID, paymentID: paymentID}

	if err := s.reserveInventory(state); err != nil {
		s.compensate(state)
		return fmt.Errorf("reserve inventory: %w", err)
	}

	if err := s.authorizePayment(state); err != nil {
		s.compensate(state)
		return fmt.Errorf("authorize payment: %w", err)
	}

	if err := s.capturePayment(state); err != nil {
		s.compensate(state)
		return fmt.Errorf("capture payment: %w", err)
	}

	if err := s.confirmOrder(state); err != nil {
		s.compensate(state)
		return fmt.Errorf("confirm order: %w", err)
	}

	return nil
}

func (s *PlaceOrderSaga) reserveInventory(state *sagaState) error {
	o, err := s.orders.FindByID(state.orderID)
	if err != nil {
		return err
	}

	for _, item := range o.Items {
		p, err := s.products.FindByID(item.ProductID)
		if err != nil {
			return err
		}
		if err := p.Reserve(state.orderID, item.Quantity); err != nil {
			return err
		}
		if err := s.products.Save(p); err != nil {
			return err
		}
		state.reservedProducts = append(state.reservedProducts, item.ProductID)
		s.bus.PublishAll(p.Events())
		p.ClearEvents()
	}
	return nil
}

func (s *PlaceOrderSaga) authorizePayment(state *sagaState) error {
	p, err := s.payments.FindByID(state.paymentID)
	if err != nil {
		return err
	}
	if err := s.gateway.Authorize(p); err != nil {
		return err
	}
	if err := s.payments.Save(p); err != nil {
		return err
	}
	s.bus.PublishAll(p.Events())
	p.ClearEvents()
	return nil
}

func (s *PlaceOrderSaga) capturePayment(state *sagaState) error {
	p, err := s.payments.FindByID(state.paymentID)
	if err != nil {
		return err
	}
	if err := s.gateway.Capture(p); err != nil {
		return err
	}
	if err := s.payments.Save(p); err != nil {
		return err
	}
	s.bus.PublishAll(p.Events())
	p.ClearEvents()
	return nil
}

func (s *PlaceOrderSaga) confirmOrder(state *sagaState) error {
	o, err := s.orders.FindByID(state.orderID)
	if err != nil {
		return err
	}

	for _, item := range o.Items {
		p, err := s.products.FindByID(item.ProductID)
		if err != nil {
			return err
		}
		if err := p.ConfirmReservation(state.orderID); err != nil {
			return err
		}
		if err := s.products.Save(p); err != nil {
			return err
		}
	}

	if err := o.Confirm(); err != nil {
		return err
	}
	if err := s.orders.Save(o); err != nil {
		return err
	}
	s.bus.PublishAll(o.Events())
	o.ClearEvents()
	return nil
}

func (s *PlaceOrderSaga) compensate(state *sagaState) {
	if state.paymentID != "" {
		p, err := s.payments.FindByID(state.paymentID)
		if err == nil && p.Status == payment.StatusAuthorized {
			_ = s.gateway.Void(p)
			_ = s.payments.Save(p)
			s.bus.PublishAll(p.Events())
			p.ClearEvents()
		}
	}

	for _, productID := range state.reservedProducts {
		p, err := s.products.FindByID(productID)
		if err == nil {
			_ = p.Release(state.orderID)
			_ = s.products.Save(p)
			s.bus.PublishAll(p.Events())
			p.ClearEvents()
		}
	}

	o, err := s.orders.FindByID(state.orderID)
	if err == nil && o.Status == order.StatusPending {
		_ = o.Cancel()
		_ = s.orders.Save(o)
		s.bus.PublishAll(o.Events())
		o.ClearEvents()
	}
}

var _ shared.DomainEvent = (*order.OrderPlaced)(nil)
