-- name: InsertOrder :exec
INSERT INTO orders (id, customer_id, status, total_amount, total_currency, placed_at)
VALUES (?, ?, ?, ?, ?, ?);

-- name: UpdateOrder :exec
UPDATE orders SET status = ?, total_amount = ?, total_currency = ? WHERE id = ?;

-- name: GetOrder :one
SELECT * FROM orders WHERE id = ?;

-- name: InsertOrderItem :exec
INSERT INTO order_items (order_id, product_id, quantity, unit_price_amount, unit_price_currency)
VALUES (?, ?, ?, ?, ?);

-- name: GetOrderItems :many
SELECT * FROM order_items WHERE order_id = ?;

-- name: DeleteOrderItems :exec
DELETE FROM order_items WHERE order_id = ?;
