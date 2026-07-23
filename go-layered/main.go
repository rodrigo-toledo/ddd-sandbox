package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/rodrigotoledo/ddd-sandbox/go-layered/handlers"
	"github.com/rodrigotoledo/ddd-sandbox/go-layered/models"
	"github.com/rodrigotoledo/ddd-sandbox/go-layered/repositories"
	"github.com/rodrigotoledo/ddd-sandbox/go-layered/services"
	_ "modernc.org/sqlite"
)

type localGateway struct{}

func (g *localGateway) Authorize(p *models.Payment) error {
	p.Status = models.PaymentAuthorized
	return nil
}
func (g *localGateway) Capture(p *models.Payment) error {
	p.Status = models.PaymentCaptured
	return nil
}
func (g *localGateway) Void(p *models.Payment) error {
	p.Status = models.PaymentVoided
	return nil
}

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
);`

func main() {
	dbPath := "go-layered.db"
	if p := os.Getenv("DB_PATH"); p != "" {
		dbPath = p
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(schema); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	orderRepo := repositories.NewOrderRepo(db)
	productRepo := repositories.NewProductRepo(db)
	paymentRepo := repositories.NewPaymentRepo(db)

	orderSvc := services.NewOrderService(orderRepo, productRepo, paymentRepo, &localGateway{})
	productSvc := services.NewProductService(productRepo)

	orderH := handlers.NewOrderHandler(orderSvc)
	productH := handlers.NewProductHandler(productSvc)
	router := handlers.NewRouter(orderH, productH)

	port := "8081"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	fmt.Printf("listening on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
