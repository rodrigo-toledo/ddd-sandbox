package service

import (
	"github.com/rodrigotoledo/ddd-sandbox/internal/application/processmanager"
	"github.com/rodrigotoledo/ddd-sandbox/internal/application/saga"
	"github.com/rodrigotoledo/ddd-sandbox/internal/domain/inventory"
	"github.com/rodrigotoledo/ddd-sandbox/internal/domain/order"
	"github.com/rodrigotoledo/ddd-sandbox/internal/domain/payment"
	"github.com/rodrigotoledo/ddd-sandbox/internal/domain/shared"
	"github.com/rodrigotoledo/ddd-sandbox/internal/infrastructure/eventbus"
)

type OrderService struct {
	orders      order.Repository
	products    inventory.Repository
	payments    payment.Repository
	bus         eventbus.EventBus
	saga        *saga.PlaceOrderSaga
	fulfillment *processmanager.FulfillmentPM
}

func NewOrderService(
	orders order.Repository,
	products inventory.Repository,
	payments payment.Repository,
	bus eventbus.EventBus,
	s *saga.PlaceOrderSaga,
	pm *processmanager.FulfillmentPM,
) *OrderService {
	return &OrderService{
		orders:      orders,
		products:    products,
		payments:    payments,
		bus:         bus,
		saga:        s,
		fulfillment: pm,
	}
}

type PlaceOrderCommand struct {
	OrderID    string
	CustomerID string
	PaymentID  string
	Items      []order.Item
}

func (s *OrderService) PlaceOrder(cmd PlaceOrderCommand) error {
	o, err := order.New(cmd.OrderID, cmd.CustomerID, cmd.Items)
	if err != nil {
		return err
	}
	if err := s.orders.Save(o); err != nil {
		return err
	}
	s.bus.PublishAll(o.Events())
	o.ClearEvents()

	p := payment.New(cmd.PaymentID, cmd.OrderID, o.Total)
	if err := s.payments.Save(p); err != nil {
		return err
	}

	if err := s.saga.Execute(cmd.OrderID, cmd.PaymentID); err != nil {
		return err
	}

	return s.fulfillment.HandleOrderConfirmed(cmd.OrderID)
}

func (s *OrderService) ShipOrder(orderID string) error {
	o, err := s.orders.FindByID(orderID)
	if err != nil {
		return err
	}
	if err := o.Ship(); err != nil {
		return err
	}
	if err := s.orders.Save(o); err != nil {
		return err
	}
	s.bus.PublishAll(o.Events())
	o.ClearEvents()
	return s.fulfillment.HandleOrderShipped(orderID)
}

func (s *OrderService) DeliverOrder(orderID string) error {
	o, err := s.orders.FindByID(orderID)
	if err != nil {
		return err
	}
	if err := o.Deliver(); err != nil {
		return err
	}
	if err := s.orders.Save(o); err != nil {
		return err
	}
	s.bus.PublishAll(o.Events())
	o.ClearEvents()
	return s.fulfillment.HandleOrderDelivered(orderID)
}

func (s *OrderService) RequestReturn(orderID string) error {
	return s.fulfillment.HandleReturnRequested(orderID)
}

var _ shared.DomainEvent = (*order.OrderPlaced)(nil)
