# Learning Guide

A reading order that maps code to DDD concepts. Read in this sequence — each step builds on the previous one.

## 1. Value Objects (start here)

**Read:** `internal/domain/shared/money.go`

A value object is defined by its attributes, not its identity. Two `Money{1000, "USD"}` are the same thing. Notice:
- No ID field
- Immutability (methods return new values, never mutate)
- Validation in the constructor (`NewMoney`)
- `Equals` compares all fields

**Ask yourself:** Why can't `Money` have a setter? What would break if two Money values with the same amount were considered different?

## 2. Aggregates & Invariants

**Read in order:**
1. `internal/domain/order/order.go`
2. `internal/domain/inventory/product.go`
3. `internal/domain/payment/payment.go`

Each aggregate root:
- Protects its **invariants** (rules that must always be true)
- Exposes behavior through methods, not field access
- Rejects invalid state transitions with errors
- Collects **domain events** as side-effects of state changes

**Key invariants to spot:**
- Order: can't ship unless confirmed, can't cancel after confirmation
- Product: available stock = stock - reservations, can't go negative
- Payment: can't capture without authorization, refund ≤ captured

**Ask yourself:** Why does `Order.Ship()` check status instead of just setting it? What would happen if external code could set `o.Status = StatusShipped` directly?

## 3. Domain Events

**Read:** `internal/domain/order/events.go`, `internal/domain/shared/event.go`

Events are facts that happened in the past. They:
- Are named in past tense (`OrderPlaced`, not `PlaceOrder`)
- Carry the data needed by consumers (no need to re-query)
- Are collected inside the aggregate (`o.events`) and published by the application layer
- Embed `BaseEvent` for common fields (name, timestamp, aggregate ID)

**Ask yourself:** Why does the aggregate collect events instead of publishing them directly? (Hint: what if the save fails after publishing?)

## 4. Repository Pattern & Dependency Inversion

**Read:**
1. `internal/domain/order/repository.go` (interface — 4 lines)
2. `internal/infrastructure/inmemory/order_repo.go` (test implementation)
3. `internal/infrastructure/persistence/order_repo.go` (SQLite implementation)

The domain defines *what* it needs (the interface). Infrastructure decides *how*. The domain never imports infrastructure.

**Ask yourself:** Why does `FindByID` return a fully reconstituted aggregate (with items) instead of a flat row? What's the "impedance mismatch" being solved here?

## 5. The Saga (short-lived coordination)

**Read:** `internal/application/saga/place_order.go`

Then read the tests: `internal/application/saga/place_order_test.go`

The saga coordinates a **distributed transaction** across aggregates:
1. Reserve inventory
2. Authorize payment
3. Capture payment
4. Confirm order

If any step fails, **compensations** run in reverse:
- Void the payment
- Release inventory
- Cancel the order

**Focus on the tests** — they show why sagas exist:
- `TestSagaHappyPath`: everything works
- `TestSagaPaymentAuthorizationFailure`: payment fails → inventory released, order cancelled
- `TestSagaCaptureFailure`: capture fails → payment voided, inventory released

**Ask yourself:** Why can't we just use a database transaction here? (Hint: in a real system, payment goes through Stripe — that's an HTTP call, not a DB operation.)

## 6. The Process Manager (long-running coordination)

**Read:** `internal/application/processmanager/fulfillment.go`

Then the tests: `internal/application/processmanager/fulfillment_test.go`

Unlike the saga (completes in milliseconds), the process manager tracks a process that spans **days or weeks**:
- Order confirmed → shipped → delivered → 30-day return window → completed

Key differences from the saga:
- **Persisted state** (survives restarts)
- **Time-aware** (return window deadline, injectable clock)
- **Reactive** (responds to events as they arrive, doesn't orchestrate calls)

**Focus on the tests:**
- `TestReturnWithinWindow`: return at day 10 → allowed
- `TestReturnWindowExpires`: 31 days pass → auto-completed
- `TestReturnRejectedAfterWindow`: return at day 31 → ignored

**Ask yourself:** Why is the clock injectable? How would you test the 30-day window without waiting 30 days?

## 7. Wiring It Together

**Read:** `internal/application/service/order_service.go`

The application service is the **use-case orchestrator**. It:
- Accepts commands (`PlaceOrderCommand`)
- Coordinates domain objects and infrastructure
- Publishes events after successful operations
- Contains no business logic itself (it delegates to aggregates and the saga)

**Then run the demo:**
```bash
go run ./cmd/demo
```

Watch the event flow in the output. Each `[event]` line is a domain event being published and consumed.

## 8. The Full Picture

Run the tests and watch them pass:
```bash
make test
```

Notice the testing pyramid:
- **Domain tests** (fast, no deps): test invariants in isolation
- **Saga/PM tests** (fast, in-memory): test coordination logic
- **Integration tests** (slower, SQLite): test persistence mapping

## Suggested Exercises

Once you've read through everything:

1. **Add a new invariant:** What if orders over $10,000 require manual approval? Where would that logic live?
2. **Add a new event:** When an order is shipped, notify the customer. Where does the notification logic go? (Hint: subscribe to `order.shipped` on the bus.)
3. **Break a test on purpose:** Remove the status check in `Order.Ship()`. Which tests fail? This shows how tests protect invariants.
4. **Trace a failure:** In `place_order_test.go`, walk through `TestSagaCaptureFailure` line by line. Draw the compensation sequence on paper.
5. **Extend the process manager:** Add a "return shipped back" state between `ReturnRequested` and `Completed`.
