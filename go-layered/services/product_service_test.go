package services

import (
	"context"
	"database/sql"
	"testing"

	"github.com/rodrigotoledo/ddd-sandbox/go-layered/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	database, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { database.Close() })

	schema := `
CREATE TABLE products (id TEXT PRIMARY KEY, name TEXT NOT NULL, stock INTEGER NOT NULL);
CREATE TABLE reservations (id INTEGER PRIMARY KEY AUTOINCREMENT, product_id TEXT NOT NULL, order_id TEXT NOT NULL, quantity INTEGER NOT NULL);
CREATE TABLE orders (id TEXT PRIMARY KEY, customer_id TEXT NOT NULL, status TEXT NOT NULL, total_amount INTEGER NOT NULL, total_currency TEXT NOT NULL, placed_at TEXT NOT NULL);
CREATE TABLE order_items (id INTEGER PRIMARY KEY AUTOINCREMENT, order_id TEXT NOT NULL, product_id TEXT NOT NULL, quantity INTEGER NOT NULL, unit_price_amount INTEGER NOT NULL, unit_price_currency TEXT NOT NULL);
CREATE TABLE payments (id TEXT PRIMARY KEY, order_id TEXT NOT NULL, amount INTEGER NOT NULL, currency TEXT NOT NULL, status TEXT NOT NULL, refunded_amount INTEGER NOT NULL DEFAULT 0);`
	_, err = database.Exec(schema)
	require.NoError(t, err)
	return database
}

func TestCreateProduct(t *testing.T) {
	database := setupTestDB(t)
	svc := NewProductService(repositories.NewProductRepo(database))

	p, err := svc.Create(context.Background(), "prod-1", "Widget", 10)
	require.NoError(t, err)
	assert.Equal(t, "prod-1", p.ID)
	assert.Equal(t, 10, p.Stock)
}

func TestGetProduct(t *testing.T) {
	database := setupTestDB(t)
	svc := NewProductService(repositories.NewProductRepo(database))

	svc.Create(context.Background(), "prod-1", "Widget", 10)
	p, err := svc.GetByID(context.Background(), "prod-1")
	require.NoError(t, err)
	assert.Equal(t, "Widget", p.Name)
}

func TestGetProductNotFound(t *testing.T) {
	database := setupTestDB(t)
	svc := NewProductService(repositories.NewProductRepo(database))

	_, err := svc.GetByID(context.Background(), "nonexistent")
	assert.ErrorIs(t, err, ErrProductNotFound)
}
