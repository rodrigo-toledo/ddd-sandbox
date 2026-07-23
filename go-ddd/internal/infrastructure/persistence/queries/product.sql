-- name: InsertProduct :exec
INSERT INTO products (id, name, stock) VALUES (?, ?, ?);

-- name: UpdateProduct :exec
UPDATE products SET name = ?, stock = ? WHERE id = ?;

-- name: GetProduct :one
SELECT * FROM products WHERE id = ?;

-- name: InsertReservation :exec
INSERT INTO reservations (product_id, order_id, quantity) VALUES (?, ?, ?);

-- name: GetReservations :many
SELECT * FROM reservations WHERE product_id = ?;

-- name: DeleteReservation :exec
DELETE FROM reservations WHERE product_id = ? AND order_id = ?;
