# DDD vs. Layered: Tradeoffs

Both implementations in this repo do the same thing — same endpoints, same behavior, same database schema. The difference is where the logic lives and what that costs you.

This isn't "DDD good, layered bad." Each has real tradeoffs. The goal here is to make those tradeoffs tangible.

## 1. Where Does the Logic Live?

**DDD:** Business rules live inside the aggregates that own the data.

```go
// go-ddd/internal/domain/order/order.go:79
func (o *Order) Ship() error {
    if o.Status != StatusConfirmed {
        return ErrNotConfirmed
    }
    o.Status = StatusShipped
    // ...
}
```

**Layered:** Business rules live in the service, separate from the data.

```go
// go-layered/services/order_service.go:148
func (s *OrderService) ShipOrder(ctx context.Context, id string) (*models.Order, error) {
    o, err := s.orders.FindByID(ctx, id)
    // ...
    if o.Status != models.OrderConfirmed {
        return nil, fmt.Errorf("%w: cannot ship order in %s status", ErrInvalidTransition, o.Status)
    }
    o.Status = models.OrderShipped
    // ...
}
```

**Tradeoff:** The DDD version makes the invariant *structural* — there is no code path that can ship an unconfirmed order, because `Ship()` is the only way to change that state. The layered version makes it *conventional* — the check exists, but any other function that accesses the repo could skip it. Today it's one check in one place. In six months, with three developers and a bulk-update script, it's a bug.

**Cost of DDD:** You can't just set a field. Every state change must go through a method, which feels ceremonial for simple cases. The layered version lets you write `o.Status = "shipped"` directly, which is faster when you're prototyping.

## 2. Coordination Across Aggregates

**DDD:** A dedicated saga struct with explicit steps and a `compensate()` method.

```go
// go-ddd/internal/application/saga/place_order.go:50
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
    // ...
}
```

**Layered:** The same logic, inline in one function.

```go
// go-layered/services/order_service.go:63
func (s *OrderService) PlaceOrder(ctx context.Context, input PlaceOrderInput) (*models.Order, error) {
    // ... 80 lines of interleaved business logic and error handling ...
    if err := s.gateway.Authorize(payment); err != nil {
        s.compensate(ctx, o, reservedProducts, payment.ID)
        return nil, fmt.Errorf("authorize payment: %w", err)
    }
    // ...
}
```

**Tradeoff:** The saga is ~190 lines across a dedicated file. The layered `PlaceOrder()` is ~100 lines in one function. The layered version is *shorter*. For this scope, that's fine.

But the saga's structure scales: each step is a named method, the compensation is in one place, and `sagaState` tracks exactly what's been done. If you add a fourth step (e.g., "notify warehouse"), you add one method and one compensation line. In the layered version, you're adding more interleaved logic to an already-long function, and you need to audit every error path to make sure compensation covers the new step.

**Cost of DDD:** The saga is a new concept to learn, a new file to maintain, and feels over-engineered for a 3-step flow. If your coordination never grows beyond 2-3 steps, the layered version is simpler.

## 3. Long-Running Processes

**DDD:** A process manager with persisted state and an injectable clock.

```go
// go-ddd/internal/application/processmanager/fulfillment.go:84
func (pm *FulfillmentPM) HandleOrderDelivered(orderID string) error {
    // ...
    now := pm.clock.Now()
    fs.State = StateDelivered
    fs.DeliveredAt = now
    fs.ReturnDeadline = now.Add(ReturnWindowDuration)
    return pm.store.Save(fs)
}
```

**Layered:** A time check inside the service method.

```go
// go-layered/services/order_service.go:178
func (s *OrderService) RequestReturn(ctx context.Context, id string, now time.Time) error {
    // ...
    deliveredAt, _ := time.Parse(time.RFC3339, o.PlacedAt)
    if now.After(deliveredAt.Add(30 * 24 * time.Hour)) {
        return ErrReturnWindowExpired
    }
    // ...
}
```

**Tradeoff:** The layered version is 5 lines. The process manager is ~140 lines. For "can I return this within 30 days?", the layered version is clearly simpler.

But the process manager earns its keep when the lifecycle grows. What if you need:
- A "return shipped back" step between request and refund?
- To track *when* the return window expires and run a background job to close it?
- To survive a server restart mid-lifecycle?

The layered version has no persisted state for the fulfillment process. It re-derives everything from the order's `PlacedAt` timestamp. The moment you need a step that doesn't map to the order's status field, you're retrofitting a state machine into a function that was never designed to hold one.

**Cost of DDD:** The process manager is infrastructure you maintain even when the lifecycle is simple. It's a state machine for a process that, today, has three states.

## 4. Testability

**DDD:** 46 tests, 43 of which need zero I/O.

```go
// go-ddd/internal/domain/order/order_test.go:56
func TestShipUnconfirmedOrder(t *testing.T) {
    o, _ := New("order-1", "cust-1", testItems())
    assert.ErrorIs(t, o.Ship(), ErrNotConfirmed)
}
```

Two lines. No database, no mocks, no setup. You can test every invariant in the domain in milliseconds.

**Layered:** 12 tests, all requiring SQLite.

```go
// go-layered/services/order_service_test.go:97
func TestShipOrderInvalidTransition(t *testing.T) {
    svc, productRepo := setupOrderTest(t)          // creates DB, repos, service
    seedProduct(t, productRepo, "prod-1", 10)      // inserts a product
    svc.PlaceOrder(context.Background(), ...)       // places an order first
    svc.ShipOrder(context.Background(), "order-1")  // ships it
    _, err := svc.ShipOrder(context.Background(), "order-1")  // tries again
    assert.ErrorIs(t, err, ErrInvalidTransition)
}
```

Seven lines, and every one of them depends on a working database. To test "can't ship twice," you first need to create a product, place an order, and ship it once.

**Tradeoff:** The DDD version has 4x more tests not because it's more complex, but because each test is cheaper to write. When testing is free, you test more. When every test needs a DB, you test the happy path and a couple of failures, then move on.

**Cost of DDD:** The in-memory repos and the repository interface are code you write and maintain purely for testability. The layered version has no interfaces, no in-memory implementations — just the real repo and a `:memory:` database.

## 5. Extensibility

Want to add a notification when an order ships?

**DDD:** Subscribe to an event. The order aggregate doesn't change.

```go
bus.Subscribe("order.shipped", func(e shared.DomainEvent) {
    // send email, push notification, whatever
})
```

**Layered:** Modify `ShipOrder()`. Or add a call after it in the handler. Or add a middleware. There's no event system, so you're coupling the notification to the shipping logic.

**Tradeoff:** Events decouple producers from consumers. But they also add indirection — when you're debugging, "who handles `order.shipped`?" requires searching for subscribers rather than reading a call stack. For a system with 2-3 consumers, the layered version's direct call is easier to trace.

**Cost of DDD:** The event bus, the event types, the `Events()`/`ClearEvents()` ceremony on every aggregate. For a system that never needs more than one consumer per action, this is overhead.

## 6. Code Volume

```
go-ddd/       ~1,400 lines of Go (excluding generated code)
go-layered/   ~700 lines of Go (excluding generated code)
```

The DDD version is roughly 2x the code. That's the tax: interfaces, in-memory implementations, event types, saga state, process manager state, value objects. You write more code to get the same behavior.

**The question is:** does the structure pay for itself as the system grows? At this scope (3 aggregates, 1 saga, 1 process manager), it's debatable. At 10 aggregates with 5 sagas and a team of 4, the DDD version's invariants, testability, and decoupling start to compound.

## When to Use Which

| Use layered when... | Use DDD when... |
|---|---|
| The domain is simple and stable | The domain has complex, evolving rules |
| You're prototyping or building a CRUD app | Multiple invariants must always hold |
| One developer, short timeline | Team of 3+, long-lived system |
| Coordination is 1-2 steps | Coordination spans multiple services/aggregates |
| You need to move fast and iterate | You need to protect against regression |

Neither is wrong. The layered version in this repo is *good code* — it's clear, it works, it's tested. DDD isn't about writing better code for its own sake. It's about managing complexity that would otherwise creep into your services until they're unmaintainable.

The best way to feel the difference: try exercise #5 from [LEARNING.md](LEARNING.md) — add a "return shipped back" state to both versions. Notice which one accommodates the change and which one fights you.
