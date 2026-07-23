package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/rodrigotoledo/ddd-sandbox/go-ddd/internal/application/processmanager"
	"github.com/rodrigotoledo/ddd-sandbox/go-ddd/internal/application/saga"
	"github.com/rodrigotoledo/ddd-sandbox/go-ddd/internal/application/service"
	"github.com/rodrigotoledo/ddd-sandbox/go-ddd/internal/domain/payment"
	"github.com/rodrigotoledo/ddd-sandbox/go-ddd/internal/domain/shared"
	apphttp "github.com/rodrigotoledo/ddd-sandbox/go-ddd/internal/infrastructure/http"
	"github.com/rodrigotoledo/ddd-sandbox/go-ddd/internal/infrastructure/eventbus"
	"github.com/rodrigotoledo/ddd-sandbox/go-ddd/internal/infrastructure/persistence"
	"github.com/rodrigotoledo/ddd-sandbox/go-ddd/internal/infrastructure/persistence/sqlc"
	_ "modernc.org/sqlite"
)

type localGateway struct{}

func (g *localGateway) Authorize(p *payment.Payment) error { return p.Authorize() }
func (g *localGateway) Capture(p *payment.Payment) error   { return p.Capture() }
func (g *localGateway) Void(p *payment.Payment) error      { return p.Void() }

const schema = `
CREATE TABLE IF NOT EXISTS orders (
    id TEXT PRIMARY KEY,
    customer_id TEXT NOT NULL,
    status TEXT NOT NULL,
    total_amount INTEGER NOT NULL,
    total_currency TEXT NOT NULL,
    placed_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS order_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    order_id TEXT NOT NULL REFERENCES orders(id),
    product_id TEXT NOT NULL,
    quantity INTEGER NOT NULL,
    unit_price_amount INTEGER NOT NULL,
    unit_price_currency TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS products (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    stock INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS reservations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    product_id TEXT NOT NULL REFERENCES products(id),
    order_id TEXT NOT NULL,
    quantity INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS payments (
    id TEXT PRIMARY KEY,
    order_id TEXT NOT NULL,
    amount INTEGER NOT NULL,
    currency TEXT NOT NULL,
    status TEXT NOT NULL,
    refunded_amount INTEGER NOT NULL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS fulfillment_states (
    order_id TEXT PRIMARY KEY,
    state TEXT NOT NULL,
    delivered_at TEXT,
    return_deadline TEXT
);`

func main() {
	dbPath := "ddd-sandbox.db"
	if p := os.Getenv("DB_PATH"); p != "" {
		dbPath = p
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if _, err := db.ExecContext(context.Background(), schema); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	q := sqlc.New(db)
	orderRepo := persistence.NewOrderRepo(q)
	productRepo := persistence.NewProductRepo(q)
	paymentRepo := persistence.NewPaymentRepo(q)
	bus := eventbus.NewInMemoryBus()
	clock := shared.SystemClock{}
	stateStore := processmanager.NewInMemoryStateStore()

	placeOrderSaga := saga.NewPlaceOrderSaga(orderRepo, productRepo, paymentRepo, &localGateway{}, bus)
	fulfillmentPM := processmanager.NewFulfillmentPM(stateStore, clock)
	svc := service.NewOrderService(orderRepo, productRepo, paymentRepo, bus, placeOrderSaga, fulfillmentPM)

	handler := apphttp.NewHandler(svc, orderRepo, productRepo)
	router := apphttp.NewRouter(handler)

	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	fmt.Printf("listening on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
