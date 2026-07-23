# DDD Sandbox

A didactic e-commerce order fulfillment application for learning Domain-Driven Design concepts, built in Go.

See [PLAN.md](PLAN.md) for scope, motivation, and architecture.

## Prerequisites

- Go 1.26+
- sqlc (`go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`)
- goose (`go install github.com/pressly/goose/v3/cmd/goose@latest`)

## Commands

```bash
make build        # Compile all packages
make test         # Run all tests
make lint         # Run go vet
make sqlc         # Generate type-safe DB code from SQL
make migrate      # Run migrations (creates ddd-sandbox.db)
make migrate-down # Rollback last migration
```

## Running the API

```bash
go run ./cmd/api
```

Then interact with it:

```bash
# Create a product
curl -X POST localhost:8080/products \
  -d '{"id":"prod-1","name":"Mechanical Keyboard","stock":50}'

# Place an order (triggers the saga: reserve → authorize → capture → confirm)
curl -X POST localhost:8080/orders \
  -d '{"order_id":"order-1","customer_id":"alice","payment_id":"pay-1","items":[{"product_id":"prod-1","quantity":2,"unit_price":15000,"currency":"USD"}]}'

# Get order status
curl localhost:8080/orders/order-1

# Ship the order
curl -X POST localhost:8080/orders/order-1/ship

# Deliver the order
curl -X POST localhost:8080/orders/order-1/deliver

# Request a return (within 30-day window)
curl -X POST localhost:8080/orders/order-1/return

# Check product stock
curl localhost:8080/products/prod-1
```

## Running the Demo (in-memory, no HTTP)

```bash
go run ./cmd/demo
```

## Project Structure

```
internal/
├── domain/           # Pure domain logic (zero external deps)
│   ├── shared/       # DomainEvent, Money, Clock
│   ├── order/        # Order aggregate
│   ├── inventory/    # Product aggregate
│   └── payment/      # Payment aggregate
├── application/      # Use-case orchestration
│   ├── saga/         # PlaceOrderSaga
│   ├── processmanager/ # FulfillmentProcessManager
│   └── service/      # Application services
└── infrastructure/   # Adapters
    ├── eventbus/     # In-memory event bus
    ├── http/         # Chi router + handlers
    ├── persistence/  # SQLite repos (hand-written mappers + sqlc-generated queries in sqlc/)
    └── inmemory/     # In-memory repos (for tests)
cmd/
├── api/              # HTTP server (SQLite-backed)
└── demo/             # In-memory lifecycle demo
```
