package saga

import (
	"errors"
	"testing"

	"github.com/rodrigotoledo/ddd-sandbox/internal/domain/inventory"
	"github.com/rodrigotoledo/ddd-sandbox/internal/domain/order"
	"github.com/rodrigotoledo/ddd-sandbox/internal/domain/payment"
	"github.com/rodrigotoledo/ddd-sandbox/internal/domain/shared"
	"github.com/rodrigotoledo/ddd-sandbox/internal/infrastructure/eventbus"
	"github.com/rodrigotoledo/ddd-sandbox/internal/infrastructure/inmemory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockGateway struct {
	authorizeErr error
	captureErr   error
}

func (g *mockGateway) Authorize(p *payment.Payment) error {
	if g.authorizeErr != nil {
		return g.authorizeErr
	}
	return p.Authorize()
}

func (g *mockGateway) Capture(p *payment.Payment) error {
	if g.captureErr != nil {
		return g.captureErr
	}
	return p.Capture()
}

func (g *mockGateway) Void(p *payment.Payment) error {
	return p.Void()
}

func setup() (*inmemory.OrderRepo, *inmemory.ProductRepo, *inmemory.PaymentRepo, *eventbus.InMemoryBus) {
	orderRepo := inmemory.NewOrderRepo()
	productRepo := inmemory.NewProductRepo()
	paymentRepo := inmemory.NewPaymentRepo()
	bus := eventbus.NewInMemoryBus()

	productRepo.Save(inventory.NewProduct("prod-1", "Widget", 10))
	productRepo.Save(inventory.NewProduct("prod-2", "Gadget", 5))

	return orderRepo, productRepo, paymentRepo, bus
}

func createTestOrder(orderRepo *inmemory.OrderRepo) *order.Order {
	o, _ := order.New("order-1", "cust-1", []order.Item{
		{ProductID: "prod-1", Quantity: 2, UnitPrice: shared.MustMoney(1000, "USD")},
		{ProductID: "prod-2", Quantity: 1, UnitPrice: shared.MustMoney(500, "USD")},
	})
	o.ClearEvents()
	orderRepo.Save(o)
	return o
}

func createTestPayment(paymentRepo *inmemory.PaymentRepo) *payment.Payment {
	p := payment.New("pay-1", "order-1", shared.MustMoney(2500, "USD"))
	paymentRepo.Save(p)
	return p
}

func TestSagaHappyPath(t *testing.T) {
	orderRepo, productRepo, paymentRepo, bus := setup()
	createTestOrder(orderRepo)
	createTestPayment(paymentRepo)

	saga := NewPlaceOrderSaga(orderRepo, productRepo, paymentRepo, &mockGateway{}, bus)
	require.NoError(t, saga.Execute("order-1", "pay-1"))

	o, _ := orderRepo.FindByID("order-1")
	assert.Equal(t, order.StatusConfirmed, o.Status)

	p1, _ := productRepo.FindByID("prod-1")
	assert.Equal(t, 8, p1.Stock)
	assert.Empty(t, p1.Reservations)

	p2, _ := productRepo.FindByID("prod-2")
	assert.Equal(t, 4, p2.Stock)

	pay, _ := paymentRepo.FindByID("pay-1")
	assert.Equal(t, payment.StatusCaptured, pay.Status)
}

func TestSagaPaymentAuthorizationFailure(t *testing.T) {
	orderRepo, productRepo, paymentRepo, bus := setup()
	createTestOrder(orderRepo)
	createTestPayment(paymentRepo)

	gw := &mockGateway{authorizeErr: errors.New("card declined")}
	saga := NewPlaceOrderSaga(orderRepo, productRepo, paymentRepo, gw, bus)

	err := saga.Execute("order-1", "pay-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "card declined")

	o, _ := orderRepo.FindByID("order-1")
	assert.Equal(t, order.StatusCancelled, o.Status)

	p1, _ := productRepo.FindByID("prod-1")
	assert.Equal(t, 10, p1.Available())
	assert.Empty(t, p1.Reservations)

	pay, _ := paymentRepo.FindByID("pay-1")
	assert.Equal(t, payment.StatusPending, pay.Status)
}

func TestSagaInsufficientStock(t *testing.T) {
	orderRepo, productRepo, paymentRepo, bus := setup()

	o, _ := order.New("order-1", "cust-1", []order.Item{
		{ProductID: "prod-1", Quantity: 99, UnitPrice: shared.MustMoney(1000, "USD")},
	})
	o.ClearEvents()
	orderRepo.Save(o)
	createTestPayment(paymentRepo)

	saga := NewPlaceOrderSaga(orderRepo, productRepo, paymentRepo, &mockGateway{}, bus)

	err := saga.Execute("order-1", "pay-1")
	require.Error(t, err)

	o, _ = orderRepo.FindByID("order-1")
	assert.Equal(t, order.StatusCancelled, o.Status)
}

func TestSagaCaptureFailure(t *testing.T) {
	orderRepo, productRepo, paymentRepo, bus := setup()
	createTestOrder(orderRepo)
	createTestPayment(paymentRepo)

	gw := &mockGateway{captureErr: errors.New("capture failed")}
	saga := NewPlaceOrderSaga(orderRepo, productRepo, paymentRepo, gw, bus)

	err := saga.Execute("order-1", "pay-1")
	require.Error(t, err)

	o, _ := orderRepo.FindByID("order-1")
	assert.Equal(t, order.StatusCancelled, o.Status)

	pay, _ := paymentRepo.FindByID("pay-1")
	assert.Equal(t, payment.StatusVoided, pay.Status)

	p1, _ := productRepo.FindByID("prod-1")
	assert.Empty(t, p1.Reservations)
}
