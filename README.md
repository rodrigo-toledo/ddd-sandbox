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
    ├── persistence/  # SQLite repos (sqlc-generated)
    └── inmemory/     # In-memory repos (for tests)
cmd/
└── demo/             # Wiring + lifecycle demo
```

## Running the Demo

```bash
go run ./cmd/demo
```
