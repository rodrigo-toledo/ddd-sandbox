-- name: InsertPayment :exec
INSERT INTO payments (id, order_id, amount, currency, status, refunded_amount)
VALUES (?, ?, ?, ?, ?, ?);

-- name: UpdatePayment :exec
UPDATE payments SET status = ?, refunded_amount = ? WHERE id = ?;

-- name: GetPayment :one
SELECT * FROM payments WHERE id = ?;

-- name: GetPaymentByOrderID :one
SELECT * FROM payments WHERE order_id = ?;
