-- +goose Up
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
);

-- +goose Down
DROP TABLE fulfillment_states;
DROP TABLE payments;
DROP TABLE reservations;
DROP TABLE products;
DROP TABLE order_items;
DROP TABLE orders;
