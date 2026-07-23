# Restructure + Layered Implementation Plan

## Step 1: Move DDD code to `go-ddd/`

- Move all files (cmd/, internal/, migrations/, Makefile, sqlc.yaml, PLAN.md, LEARNING.md) into `go-ddd/`
- Update module path: `github.com/rodrigotoledo/ddd-sandbox/go-ddd`
- Fix all import paths
- **Verify:** `make build && make test && make lint` in `go-ddd/`
- **Commit:** `restructure: move ddd implementation to go-ddd/`

## Step 2: Scaffold `go-layered/`

- `go.mod` (module `github.com/rodrigotoledo/ddd-sandbox/go-layered`)
- Makefile (build, test, lint, sqlc)
- Directory structure: `models/`, `services/`, `handlers/`, `repositories/`, `migrations/`
- Minimal `main.go` that compiles
- **Verify:** `make build && make lint`
- **Commit:** `go-layered: scaffold project structure`

## Step 3: Models (anemic structs)

- `models/order.go` — `Order`, `OrderItem`, status constants
- `models/product.go` — `Product`, `Reservation`
- `models/payment.go` — `Payment`, status constants
- No methods, no behavior — just fields
- **Verify:** `make build`
- **Commit:** `go-layered: anemic models (order, product, payment)`

## Step 4: Migrations + sqlc

- `migrations/001_init.sql` — same schema as DDD version
- `sqlc.yaml` + query files (same queries)
- Run `sqlc generate`
- **Verify:** `make build && make lint`
- **Commit:** `go-layered: migrations + sqlc generated code`

## Step 5: Repositories

- `repositories/order_repo.go` — Save, FindByID (maps sqlc rows ↔ models)
- `repositories/product_repo.go` — Save, FindByID
- `repositories/payment_repo.go` — Save, FindByID, FindByOrderID
- **Verify:** `make build && make lint`
- **Commit:** `go-layered: sqlc-backed repositories`

## Step 6: Product service + tests

- `services/product_service.go` — Create, GetByID
- `services/product_service_test.go` — create product, get product, not found
- Tests use SQLite `:memory:`
- **Verify:** `make test && make lint`
- **Commit:** `go-layered: product service with tests`

## Step 7: Order service — PlaceOrder + tests

- `services/order_service.go` — `PlaceOrder()`:
  - Validate items, compute total
  - Insert order (status: pending)
  - For each item: check stock, reserve
  - Create payment, authorize, capture
  - Confirm order, decrement stock
  - On failure: inline compensation (release stock, void payment, cancel order)
- `services/order_service_test.go`:
  - Happy path (order confirmed, stock decremented, payment captured)
  - Insufficient stock (order cancelled, nothing reserved)
  - Payment failure (order cancelled, stock released)
- **Verify:** `make test && make lint`
- **Commit:** `go-layered: place order with inline compensation + tests`

## Step 8: Order service — Ship, Deliver, Return + tests

- Add to `services/order_service.go`:
  - `ShipOrder()` — check status == confirmed, update
  - `DeliverOrder()` — check status == shipped, update, record delivered_at
  - `RequestReturn()` — check status == delivered, check return window
  - `CheckReturnWindow()` — expire old returns (the "process manager" equivalent)
- Add tests:
  - Ship happy path + invalid transition
  - Deliver happy path + invalid transition
  - Return within window + after window expired
- **Verify:** `make test && make lint`
- **Commit:** `go-layered: ship/deliver/return with status checks + tests`

## Step 9: HTTP handlers + router

- `handlers/order.go` — same request/response DTOs as DDD version
- `handlers/product.go` — same
- `handlers/router.go` — chi router, same routes, same middleware
- **Verify:** `make build && make lint`
- **Commit:** `go-layered: http handlers + chi router`

## Step 10: main.go wiring

- Open SQLite, run schema
- Wire repos → services → handlers → router
- Start server on :8080
- **Verify:** start server, curl full lifecycle (create product → place → ship → deliver → return)
- **Commit:** `go-layered: server wiring, full lifecycle works`

## Step 11: Root README + final pass

- Root `README.md`: explains both implementations, how to run each, what to compare
- Update `.gitignore`
- **Verify:** both `make test` pass, both servers respond to same curl commands
- **Commit:** `docs: root readme comparing both architectures`

## Comparison Guide

After both are done, the interesting diffs:

```bash
# The "saga" — DDD makes compensations explicit:
go-ddd/internal/application/saga/place_order.go    # ~140 lines, structured

# vs. layered — it's inline in the service:
go-layered/services/order_service.go               # one long PlaceOrder() function

# The "process manager" — DDD has a state machine:
go-ddd/internal/application/processmanager/fulfillment.go

# vs. layered — it's a status check:
go-layered/services/order_service.go               # CheckReturnWindow() function

# Testing — DDD tests are pure:
go-ddd/internal/application/saga/place_order_test.go  # in-memory, no DB

# vs. layered — tests need SQLite:
go-layered/services/order_service_test.go             # :memory: DB setup
```
