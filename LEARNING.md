# Learning Guide

A reading order that maps code to DDD concepts. Each section shows you the DDD implementation, then contrasts it with the layered version so you can feel *why* the DDD approach is structured that way.

Read in sequence — each step builds on the previous one.

## 1. Value Objects (start here)

**Read:** `go-ddd/internal/domain/shared/money.go`

A value object is defined by its attributes, not its identity. Two `Money{1000, "USD"}` are the same thing. Notice:
- No ID field
- Immutability (methods return new values, never mutate)
- Validation in the constructor (`NewMoney`)
- `Equals` compares all fields

**Now compare:** The layered version has no Money type at all — prices are bare `int64` fields scattered across structs (`go-layered/models/order.go`). Nothing stops you from adding a USD amount to a EUR amount, or passing a negative price. The value object makes invalid states unrepresentable.

**Ask yourself:** Why can't `Money` have a setter? What would break if two Money values with the same amount were considered different?

## 2. Aggregates & Invariants

**Read in order:**
1. `go-ddd/internal/domain/order/order.go`
2. `go-ddd/internal/domain/inventory/product.go`
3. `go-ddd/internal/domain/payment/payment.go`

Each aggregate root:
- Protects its **invariants** (rules that must always be true)
- Exposes behavior through methods, not field access
- Rejects invalid state transitions with errors
- Collects **domain events** as side-effects of state changes

**Key invariants to spot:**
- Order: can't ship unless confirmed, can't cancel after confirmation
- Product: available stock = stock - reservations, can't go negative
- Payment: can't capture without authorization, refund ≤ captured

**Now compare:** Open `go-layered/services/order_service.go` and find `ShipOrder()`. The same invariant exists — `if o.Status != models.OrderConfirmed` — but it lives in the *service*, not the model. Any other code path that touches the order can skip this check. In the DDD version, there is no way to ship an unconfirmed order because `Order.Ship()` is the *only* way to change that state.

**Ask yourself:** In the layered version, what happens if a new developer writes a bulk-update script that sets `status = "shipped"` directly in the DB? How does the DDD version protect against that?

## 3. Domain Events

**Read:** `go-ddd/internal/domain/order/events.go`, `go-ddd/internal/domain/shared/event.go`

Events are facts that happened in the past. They:
- Are named in past tense (`OrderPlaced`, not `PlaceOrder`)
- Carry the data needed by consumers (no need to re-query)
- Are collected inside the aggregate (`o.events`) and published by the application layer
- Embed `BaseEvent` for common fields (name, timestamp, aggregate ID)

**Now compare:** The layered version has no events at all. When an order is placed, nothing *announces* it. If you later need to send a confirmation email, update a search index, or notify a warehouse — you'd have to find the `PlaceOrder()` function and add more code to it. In the DDD version, you just subscribe to `order.confirmed` on the bus. The order aggregate doesn't know or care who's listening.

**Ask yourself:** Why does the aggregate collect events instead of publishing them directly? (Hint: what if the save fails after publishing?)

## 4. Repository Pattern & Dependency Inversion

**Read:**
1. `go-ddd/internal/domain/order/repository.go` (interface — just `Save` and `FindByID`)
2. `go-ddd/internal/infrastructure/inmemory/order_repo.go` (test implementation)
3. `go-ddd/internal/infrastructure/persistence/order_repo.go` (SQLite implementation)

The domain defines *what* it needs (the interface). Infrastructure decides *how*. The domain never imports infrastructure.

**Now compare:** In `go-layered/repositories/order_repo.go`, the repository is a concrete struct that the service depends on directly. To test the service, you *must* have a database (even if it's `:memory:`). In the DDD version, the saga tests use `inmemory.OrderRepo` — a 15-line map — and run in microseconds with zero I/O.

**Ask yourself:** Why does `FindByID` return a fully reconstituted aggregate (with items) instead of a flat row? What's the "impedance mismatch" being solved here?

## 5. The Saga (short-lived coordination)

**Read:** `go-ddd/internal/application/saga/place_order.go`

Then read the tests: `go-ddd/internal/application/saga/place_order_test.go`

The saga coordinates a **distributed transaction** across aggregates:
1. Reserve inventory
2. Authorize payment
3. Capture payment
4. Confirm order

If any step fails, **compensations** run in reverse:
- Void the payment
- Release inventory
- Cancel the order

**Now compare:** Open `go-layered/services/order_service.go` and read `PlaceOrder()`. It does the *exact same thing* — same steps, same compensations. But it's all inline in one 80-line function. The compensation logic (`s.compensate(...)`) is called from multiple error paths, and you have to mentally trace which products were already reserved at each failure point. The saga makes the steps and compensations *structural* — you can read the `Execute()` method as a checklist.

**Focus on the tests** — they show why sagas exist:
- `TestSagaHappyPath`: everything works
- `TestSagaPaymentAuthorizationFailure`: payment fails → inventory released, order cancelled
- `TestSagaCaptureFailure`: capture fails → payment voided, inventory released

**Ask yourself:** Why can't we just use a database transaction here? (Hint: in a real system, payment goes through Stripe — that's an HTTP call, not a DB operation.)

## 6. The Process Manager (long-running coordination)

**Read:** `go-ddd/internal/application/processmanager/fulfillment.go`

Then the tests: `go-ddd/internal/application/processmanager/fulfillment_test.go`

Unlike the saga (completes in milliseconds), the process manager tracks a process that spans **days or weeks**:
- Order confirmed → shipped → delivered → 30-day return window → completed

Key differences from the saga:
- **Persisted state** (survives restarts)
- **Time-aware** (return window deadline, injectable clock)
- **Invoked by the application service** after each domain event (not subscribed to the bus directly — a deliberate simplification; in production you'd wire it as an event subscriber)

**Now compare:** In `go-layered/services/order_service.go`, find `RequestReturn()`. The "process manager" is just an `if` check: `if now.After(deliveredAt.Add(30 * 24 * time.Hour))`. There's no persisted state tracking where in the lifecycle we are. If you need to add a "return shipped back" step, or track partial refunds, or handle a "return denied" path — you're retrofitting state into a function that was never designed to hold it. The process manager makes the lifecycle *explicit* and *extensible*.

**Focus on the tests:**
- `TestReturnWithinWindow`: return at day 10 → allowed
- `TestReturnWindowExpires`: 31 days pass → auto-completed
- `TestReturnRejectedAfterWindow`: return at day 31 → ignored

**Ask yourself:** Why is the clock injectable? How would you test the 30-day window without waiting 30 days? (Now look at the layered test — it passes `time.Now().Add(31*24*time.Hour)` as a parameter. Same trick, but the DDD version makes it a first-class dependency.)

## 7. Wiring It Together

**Read:** `go-ddd/internal/application/service/order_service.go`

The application service is the **use-case orchestrator**. It:
- Accepts commands (`PlaceOrderCommand`)
- Coordinates domain objects and infrastructure
- Publishes events after successful operations
- Contains no business logic itself (it delegates to aggregates and the saga)

**Now compare:** `go-layered/services/order_service.go` is also called a "service" — but it contains *all* the business logic. Validation, state transitions, compensation, time checks — everything lives here. The DDD application service is thin by design: it's glue, not logic.

**Run both demos:**
```bash
cd go-ddd && go run ./cmd/demo
cd go-layered && go run .  # then curl the endpoints
```

## 8. The Full Picture

Run the tests and compare:
```bash
cd go-ddd && make test      # 46 tests, most are pure (no DB, no I/O)
cd go-layered && make test  # 12 tests, all need SQLite :memory:
```

Notice the testing pyramid:
- **DDD domain tests** (fast, no deps): test invariants in isolation
- **DDD saga/PM tests** (fast, in-memory): test coordination logic
- **DDD integration tests** (slower, SQLite): test persistence mapping
- **Layered tests** (all integration): every test spins up a database

The DDD version has more tests *because* the logic is easier to test. The layered version has fewer tests not because it's simpler, but because each test is heavier to write.

## Suggested Exercises

Once you've read through everything:

1. **Add a new invariant:** What if orders over $10,000 require manual approval? Add it to both versions. Notice where the logic lives in each.
2. **Add a new event subscriber:** When an order is shipped, log a notification. In the DDD version, subscribe to `order.shipped`. In the layered version... where do you put it?
3. **Break a test on purpose:** Remove the status check in `go-ddd/internal/domain/order/order.go` `Ship()`. Which tests fail? Now remove the same check in `go-layered/services/order_service.go`. Which tests fail? (Hint: the DDD version catches it at the domain level.)
4. **Trace a failure:** In both versions, walk through what happens when payment capture fails. Draw the compensation sequence on paper. Which one is easier to follow?
5. **Extend the process manager:** Add a "return shipped back" state between `ReturnRequested` and `Completed` in the DDD version. Now try the same in the layered version. Which one accommodates the change more gracefully?
