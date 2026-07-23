package persistence

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/rodrigotoledo/ddd-sandbox/internal/domain/inventory"
	"github.com/rodrigotoledo/ddd-sandbox/internal/domain/order"
	"github.com/rodrigotoledo/ddd-sandbox/internal/domain/payment"
	"github.com/rodrigotoledo/ddd-sandbox/internal/domain/shared"
	"github.com/rodrigotoledo/ddd-sandbox/internal/infrastructure/persistence/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func setupDB(t *testing.T) *sqlc.Queries {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	schema := `
CREATE TABLE orders (
    id TEXT PRIMARY KEY,
    customer_id TEXT NOT NULL,
    status TEXT NOT NULL,
    total_amount INTEGER NOT NULL,
    total_currency TEXT NOT NULL,
    placed_at TEXT NOT NULL
);
CREATE TABLE order_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    order_id TEXT NOT NULL REFERENCES orders(id),
    product_id TEXT NOT NULL,
    quantity INTEGER NOT NULL,
    unit_price_amount INTEGER NOT NULL,
    unit_price_currency TEXT NOT NULL
);
CREATE TABLE products (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    stock INTEGER NOT NULL
);
CREATE TABLE reservations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    product_id TEXT NOT NULL REFERENCES products(id),
    order_id TEXT NOT NULL,
    quantity INTEGER NOT NULL
);
CREATE TABLE payments (
    id TEXT PRIMARY KEY,
    order_id TEXT NOT NULL,
    amount INTEGER NOT NULL,
    currency TEXT NOT NULL,
    status TEXT NOT NULL,
    refunded_amount INTEGER NOT NULL DEFAULT 0
);
CREATE TABLE fulfillment_states (
    order_id TEXT PRIMARY KEY,
    state TEXT NOT NULL,
    delivered_at TEXT,
    return_deadline TEXT
);`
	_, err = db.ExecContext(ctx, schema)
	require.NoError(t, err)

	return sqlc.New(db)
}

func TestOrderRepoPersistence(t *testing.T) {
	q := setupDB(t)
	repo := NewOrderRepo(q)

	o, err := order.New("order-1", "cust-1", []order.Item{
		{ProductID: "prod-1", Quantity: 2, UnitPrice: shared.MustMoney(1000, "USD")},
		{ProductID: "prod-2", Quantity: 1, UnitPrice: shared.MustMoney(500, "USD")},
	})
	require.NoError(t, err)
	require.NoError(t, repo.Save(o))

	found, err := repo.FindByID("order-1")
	require.NoError(t, err)
	assert.Equal(t, "order-1", found.ID)
	assert.Equal(t, "cust-1", found.CustomerID)
	assert.Equal(t, order.StatusPending, found.Status)
	assert.Equal(t, int64(2500), found.Total.Amount)
	assert.Len(t, found.Items, 2)

	require.NoError(t, found.Confirm())
	require.NoError(t, repo.Save(found))

	updated, err := repo.FindByID("order-1")
	require.NoError(t, err)
	assert.Equal(t, order.StatusConfirmed, updated.Status)
}

func TestProductRepoPersistence(t *testing.T) {
	q := setupDB(t)
	repo := NewProductRepo(q)

	p := inventory.NewProduct("prod-1", "Widget", 10)
	require.NoError(t, repo.Save(p))

	found, err := repo.FindByID("prod-1")
	require.NoError(t, err)
	assert.Equal(t, "Widget", found.Name)
	assert.Equal(t, 10, found.Stock)

	require.NoError(t, found.Reserve("order-1", 3))
	require.NoError(t, repo.Save(found))

	updated, err := repo.FindByID("prod-1")
	require.NoError(t, err)
	assert.Equal(t, 7, updated.Available())
	assert.Len(t, updated.Reservations, 1)

	require.NoError(t, updated.Release("order-1"))
	require.NoError(t, repo.Save(updated))

	released, err := repo.FindByID("prod-1")
	require.NoError(t, err)
	assert.Equal(t, 10, released.Available())
	assert.Empty(t, released.Reservations)
}

func TestPaymentRepoPersistence(t *testing.T) {
	q := setupDB(t)
	repo := NewPaymentRepo(q)

	p := payment.New("pay-1", "order-1", shared.MustMoney(2500, "USD"))
	require.NoError(t, repo.Save(p))

	found, err := repo.FindByID("pay-1")
	require.NoError(t, err)
	assert.Equal(t, payment.StatusPending, found.Status)

	require.NoError(t, found.Authorize())
	require.NoError(t, repo.Save(found))

	updated, err := repo.FindByOrderID("order-1")
	require.NoError(t, err)
	assert.Equal(t, payment.StatusAuthorized, updated.Status)
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
