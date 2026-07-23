package main

import (
	"fmt"
	"log"

	"github.com/rodrigotoledo/ddd-sandbox/go-ddd/internal/application/processmanager"
	"github.com/rodrigotoledo/ddd-sandbox/go-ddd/internal/application/saga"
	"github.com/rodrigotoledo/ddd-sandbox/go-ddd/internal/application/service"
	"github.com/rodrigotoledo/ddd-sandbox/go-ddd/internal/domain/inventory"
	"github.com/rodrigotoledo/ddd-sandbox/go-ddd/internal/domain/order"
	"github.com/rodrigotoledo/ddd-sandbox/go-ddd/internal/domain/payment"
	"github.com/rodrigotoledo/ddd-sandbox/go-ddd/internal/domain/shared"
	"github.com/rodrigotoledo/ddd-sandbox/go-ddd/internal/infrastructure/eventbus"
	"github.com/rodrigotoledo/ddd-sandbox/go-ddd/internal/infrastructure/inmemory"
)

type localGateway struct{}

func (g *localGateway) Authorize(p *payment.Payment) error { return p.Authorize() }
func (g *localGateway) Capture(p *payment.Payment) error   { return p.Capture() }
func (g *localGateway) Void(p *payment.Payment) error      { return p.Void() }

func main() {
	orderRepo := inmemory.NewOrderRepo()
	productRepo := inmemory.NewProductRepo()
	paymentRepo := inmemory.NewPaymentRepo()
	bus := eventbus.NewInMemoryBus()
	clock := shared.SystemClock{}
	stateStore := processmanager.NewInMemoryStateStore()

	productRepo.Save(inventory.NewProduct("prod-1", "Mechanical Keyboard", 50))
	productRepo.Save(inventory.NewProduct("prod-2", "USB-C Cable", 200))

	bus.Subscribe("order.confirmed", func(e shared.DomainEvent) {
		fmt.Printf("  [event] %s (order=%s)\n", e.EventName(), e.AggregateID())
	})
	bus.Subscribe("order.shipped", func(e shared.DomainEvent) {
		fmt.Printf("  [event] %s (order=%s)\n", e.EventName(), e.AggregateID())
	})
	bus.Subscribe("order.delivered", func(e shared.DomainEvent) {
		fmt.Printf("  [event] %s (order=%s)\n", e.EventName(), e.AggregateID())
	})
	bus.Subscribe("inventory.reserved", func(e shared.DomainEvent) {
		fmt.Printf("  [event] %s (product=%s)\n", e.EventName(), e.AggregateID())
	})
	bus.Subscribe("payment.authorized", func(e shared.DomainEvent) {
		fmt.Printf("  [event] %s (payment=%s)\n", e.EventName(), e.AggregateID())
	})
	bus.Subscribe("payment.captured", func(e shared.DomainEvent) {
		fmt.Printf("  [event] %s (payment=%s)\n", e.EventName(), e.AggregateID())
	})

	placeOrderSaga := saga.NewPlaceOrderSaga(orderRepo, productRepo, paymentRepo, &localGateway{}, bus)
	fulfillmentPM := processmanager.NewFulfillmentPM(stateStore, clock)
	svc := service.NewOrderService(orderRepo, productRepo, paymentRepo, bus, placeOrderSaga, fulfillmentPM)

	fmt.Println("=== Placing order ===")
	err := svc.PlaceOrder(service.PlaceOrderCommand{
		OrderID:    "order-001",
		CustomerID: "customer-alice",
		PaymentID:  "pay-001",
		Items: []order.Item{
			{ProductID: "prod-1", Quantity: 1, UnitPrice: shared.MustMoney(15000, "USD")},
			{ProductID: "prod-2", Quantity: 2, UnitPrice: shared.MustMoney(1200, "USD")},
		},
	})
	if err != nil {
		log.Fatalf("place order failed: %v", err)
	}

	o, _ := orderRepo.FindByID("order-001")
	fmt.Printf("  order status: %s, total: %s\n", o.Status, o.Total)

	fmt.Println("\n=== Shipping order ===")
	if err := svc.ShipOrder("order-001"); err != nil {
		log.Fatalf("ship failed: %v", err)
	}
	o, _ = orderRepo.FindByID("order-001")
	fmt.Printf("  order status: %s\n", o.Status)

	fmt.Println("\n=== Delivering order ===")
	if err := svc.DeliverOrder("order-001"); err != nil {
		log.Fatalf("deliver failed: %v", err)
	}
	o, _ = orderRepo.FindByID("order-001")
	fmt.Printf("  order status: %s\n", o.Status)

	fs, _ := stateStore.FindByOrderID("order-001")
	fmt.Printf("  fulfillment state: %s, return deadline: %s\n", fs.State, fs.ReturnDeadline.Format("2006-01-02"))

	fmt.Println("\n=== Done ===")
}
