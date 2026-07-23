package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rodrigotoledo/ddd-sandbox/go-layered/models"
	"github.com/rodrigotoledo/ddd-sandbox/go-layered/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockGateway struct {
	authorizeErr error
	captureErr   error
}

func (g *mockGateway) Authorize(p *models.Payment) error {
	if g.authorizeErr != nil {
		return g.authorizeErr
	}
	p.Status = models.PaymentAuthorized
	return nil
}

func (g *mockGateway) Capture(p *models.Payment) error {
	if g.captureErr != nil {
		return g.captureErr
	}
	p.Status = models.PaymentCaptured
	return nil
}

func (g *mockGateway) Void(p *models.Payment) error {
	p.Status = models.PaymentVoided
	return nil
}

func setupOrderTest(t *testing.T) (*OrderService, *repositories.ProductRepo) {
	t.Helper()
	database := setupTestDB(t)
	orderRepo := repositories.NewOrderRepo(database)
	productRepo := repositories.NewProductRepo(database)
	paymentRepo := repositories.NewPaymentRepo(database)
	svc := NewOrderService(orderRepo, productRepo, paymentRepo, &mockGateway{})
	return svc, productRepo
}

func seedProduct(t *testing.T, repo *repositories.ProductRepo, id string, stock int) {
	t.Helper()
	require.NoError(t, repo.Save(context.Background(), &models.Product{ID: id, Name: "Test", Stock: stock}))
}

func TestPlaceOrderHappyPath(t *testing.T) {
	svc, productRepo := setupOrderTest(t)
	seedProduct(t, productRepo, "prod-1", 10)

	o, err := svc.PlaceOrder(context.Background(), PlaceOrderInput{
		OrderID:    "order-1",
		CustomerID: "cust-1",
		PaymentID:  "pay-1",
		Items:      []models.OrderItem{{ProductID: "prod-1", Quantity: 2, UnitPrice: 1000, Currency: "USD"}},
	})
	require.NoError(t, err)
	assert.Equal(t, models.OrderConfirmed, o.Status)
	assert.Equal(t, int64(2000), o.Total)

	p, _ := productRepo.FindByID(context.Background(), "prod-1")
	assert.Equal(t, 8, p.Stock)
}

func TestPlaceOrderInsufficientStock(t *testing.T) {
	svc, productRepo := setupOrderTest(t)
	seedProduct(t, productRepo, "prod-1", 1)

	_, err := svc.PlaceOrder(context.Background(), PlaceOrderInput{
		OrderID:    "order-1",
		CustomerID: "cust-1",
		PaymentID:  "pay-1",
		Items:      []models.OrderItem{{ProductID: "prod-1", Quantity: 5, UnitPrice: 1000, Currency: "USD"}},
	})
	require.ErrorIs(t, err, ErrInsufficientStock)

	o, _ := svc.GetOrder(context.Background(), "order-1")
	assert.Equal(t, models.OrderCancelled, o.Status)
}

func TestPlaceOrderPaymentFailure(t *testing.T) {
	database := setupTestDB(t)
	orderRepo := repositories.NewOrderRepo(database)
	productRepo := repositories.NewProductRepo(database)
	paymentRepo := repositories.NewPaymentRepo(database)
	seedProduct(t, productRepo, "prod-1", 10)

	gw := &mockGateway{authorizeErr: errors.New("card declined")}
	svc := NewOrderService(orderRepo, productRepo, paymentRepo, gw)

	_, err := svc.PlaceOrder(context.Background(), PlaceOrderInput{
		OrderID:    "order-1",
		CustomerID: "cust-1",
		PaymentID:  "pay-1",
		Items:      []models.OrderItem{{ProductID: "prod-1", Quantity: 2, UnitPrice: 1000, Currency: "USD"}},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "card declined")

	o, _ := svc.GetOrder(context.Background(), "order-1")
	assert.Equal(t, models.OrderCancelled, o.Status)

	p, _ := productRepo.FindByID(context.Background(), "prod-1")
	assert.Equal(t, 10, p.Stock)
}

func TestShipOrder(t *testing.T) {
	svc, productRepo := setupOrderTest(t)
	seedProduct(t, productRepo, "prod-1", 10)

	svc.PlaceOrder(context.Background(), PlaceOrderInput{
		OrderID: "order-1", CustomerID: "cust-1", PaymentID: "pay-1",
		Items: []models.OrderItem{{ProductID: "prod-1", Quantity: 1, UnitPrice: 500, Currency: "USD"}},
	})

	o, err := svc.ShipOrder(context.Background(), "order-1")
	require.NoError(t, err)
	assert.Equal(t, models.OrderShipped, o.Status)
}

func TestShipOrderInvalidTransition(t *testing.T) {
	svc, productRepo := setupOrderTest(t)
	seedProduct(t, productRepo, "prod-1", 10)

	svc.PlaceOrder(context.Background(), PlaceOrderInput{
		OrderID: "order-1", CustomerID: "cust-1", PaymentID: "pay-1",
		Items: []models.OrderItem{{ProductID: "prod-1", Quantity: 1, UnitPrice: 500, Currency: "USD"}},
	})
	svc.ShipOrder(context.Background(), "order-1")

	_, err := svc.ShipOrder(context.Background(), "order-1")
	assert.ErrorIs(t, err, ErrInvalidTransition)
}

func TestDeliverOrder(t *testing.T) {
	svc, productRepo := setupOrderTest(t)
	seedProduct(t, productRepo, "prod-1", 10)

	svc.PlaceOrder(context.Background(), PlaceOrderInput{
		OrderID: "order-1", CustomerID: "cust-1", PaymentID: "pay-1",
		Items: []models.OrderItem{{ProductID: "prod-1", Quantity: 1, UnitPrice: 500, Currency: "USD"}},
	})
	svc.ShipOrder(context.Background(), "order-1")

	o, err := svc.DeliverOrder(context.Background(), "order-1")
	require.NoError(t, err)
	assert.Equal(t, models.OrderDelivered, o.Status)
}

func TestDeliverOrderInvalidTransition(t *testing.T) {
	svc, productRepo := setupOrderTest(t)
	seedProduct(t, productRepo, "prod-1", 10)

	svc.PlaceOrder(context.Background(), PlaceOrderInput{
		OrderID: "order-1", CustomerID: "cust-1", PaymentID: "pay-1",
		Items: []models.OrderItem{{ProductID: "prod-1", Quantity: 1, UnitPrice: 500, Currency: "USD"}},
	})

	_, err := svc.DeliverOrder(context.Background(), "order-1")
	assert.ErrorIs(t, err, ErrInvalidTransition)
}

func TestRequestReturnWithinWindow(t *testing.T) {
	svc, productRepo := setupOrderTest(t)
	seedProduct(t, productRepo, "prod-1", 10)

	svc.PlaceOrder(context.Background(), PlaceOrderInput{
		OrderID: "order-1", CustomerID: "cust-1", PaymentID: "pay-1",
		Items: []models.OrderItem{{ProductID: "prod-1", Quantity: 1, UnitPrice: 500, Currency: "USD"}},
	})
	svc.ShipOrder(context.Background(), "order-1")
	svc.DeliverOrder(context.Background(), "order-1")

	err := svc.RequestReturn(context.Background(), "order-1", time.Now())
	assert.NoError(t, err)
}

func TestRequestReturnAfterWindow(t *testing.T) {
	svc, productRepo := setupOrderTest(t)
	seedProduct(t, productRepo, "prod-1", 10)

	svc.PlaceOrder(context.Background(), PlaceOrderInput{
		OrderID: "order-1", CustomerID: "cust-1", PaymentID: "pay-1",
		Items: []models.OrderItem{{ProductID: "prod-1", Quantity: 1, UnitPrice: 500, Currency: "USD"}},
	})
	svc.ShipOrder(context.Background(), "order-1")
	svc.DeliverOrder(context.Background(), "order-1")

	err := svc.RequestReturn(context.Background(), "order-1", time.Now().Add(31*24*time.Hour))
	assert.ErrorIs(t, err, ErrReturnWindowExpired)
}
