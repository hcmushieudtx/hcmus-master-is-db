-- name: CreateOrder :one
INSERT INTO orders (user_id, status, total_amount, shipping_address, payment_method)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetOrderByID :one
SELECT * FROM orders WHERE id = $1 LIMIT 1;

-- name: ListOrdersByUser :many
SELECT * FROM orders
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountOrdersByUser :one
SELECT COUNT(*) FROM orders WHERE user_id = $1;

-- name: ListAllOrders :many
SELECT * FROM orders
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListOrdersByStatus :many
SELECT * FROM orders
WHERE status = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountAllOrders :one
SELECT COUNT(*) FROM orders;

-- name: UpdateOrderStatus :one
UPDATE orders
SET status     = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CreateOrderItem :one
INSERT INTO order_items (order_id, mongo_book_id, title, quantity, unit_price)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetOrderItems :many
SELECT * FROM order_items WHERE order_id = $1;

-- name: GetBookRef :one
SELECT * FROM books_ref WHERE mongo_id = $1 LIMIT 1;

-- name: GetBookRefForUpdate :one
SELECT * FROM books_ref WHERE mongo_id = $1 LIMIT 1 FOR UPDATE;

-- name: DecrementStock :one
UPDATE books_ref
SET stock_qty = stock_qty - $2
WHERE mongo_id = $1 AND stock_qty >= $2
RETURNING *;

-- name: SumRevenueByDateRange :one
SELECT COALESCE(SUM(total_amount), 0) AS revenue,
       COUNT(*)                        AS total_orders
FROM orders
WHERE created_at BETWEEN $1 AND $2
  AND status != 'cancelled';
