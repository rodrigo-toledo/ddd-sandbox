# DDD Sandbox

The same e-commerce order fulfillment application implemented two ways — **DDD** and **traditional layered architecture** — so you can compare them side by side.

Same domain, same API endpoints, same behavior. The difference is how the logic is organized.

## Running

Both servers expose identical endpoints. The DDD version runs on `:8080`, the layered on `:8081`.

```bash
# DDD version
cd go-ddd && go run ./cmd/api

# Layered version
cd go-layered && go run .
```

## API (identical for both)

```bash
curl -X POST localhost:8080/products \
  -d '{"id":"prod-1","name":"Mechanical Keyboard","stock":50}'

curl -X POST localhost:8080/orders \
  -d '{"order_id":"order-1","customer_id":"alice","payment_id":"pay-1","items":[{"product_id":"prod-1","quantity":2,"unit_price":15000,"currency":"USD"}]}'

curl localhost:8080/orders/order-1
curl -X POST localhost:8080/orders/order-1/ship
curl -X POST localhost:8080/orders/order-1/deliver
curl -X POST localhost:8080/orders/order-1/return
curl localhost:8080/products/prod-1
```

## What to Compare

| Concern | go-ddd | go-layered |
|---|---|---|
| Business logic | Inside aggregates (methods guard invariants) | In service functions (`if` checks) |
| Models | Rich (behavior + state together) | Anemic (fields only) |
| Place order coordination | `PlaceOrderSaga` with explicit `compensate()` | One long `PlaceOrder()` with inline rollback |
| Fulfillment tracking | `FulfillmentPM` state machine + injectable clock | Status field + time check in `RequestReturn()` |
| Domain events | Explicit events on a bus | None — direct calls |
| Testing | Pure unit tests (in-memory repos, no DB) | Integration tests (SQLite `:memory:`) |
| Invariant enforcement | Compiler-enforced (can't bypass aggregate methods) | Convention-enforced (any code can set fields) |

## Key Files to Diff

```bash
# The "saga" vs. inline compensation
go-ddd/internal/application/saga/place_order.go
go-layered/services/order_service.go

# The "process manager" vs. status check
go-ddd/internal/application/processmanager/fulfillment.go
go-layered/services/order_service.go  # RequestReturn()

# Rich model vs. anemic model
go-ddd/internal/domain/order/order.go
go-layered/models/order.go

# Pure tests vs. DB-dependent tests
go-ddd/internal/application/saga/place_order_test.go
go-layered/services/order_service_test.go
```

## Development

```bash
# Each is an independent Go module
cd go-ddd && make test
cd go-layered && make test
```

## Docs

- [go-ddd/PLAN.md](go-ddd/PLAN.md) — DDD scope, architecture, growth path
- [go-ddd/LEARNING.md](go-ddd/LEARNING.md) — guided reading order for DDD concepts
- [RESTRUCTURE_PLAN.md](RESTRUCTURE_PLAN.md) — implementation plan for this restructure
