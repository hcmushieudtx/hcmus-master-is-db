-- name: CreateUser :one
INSERT INTO users (full_name, email, phone, password_hash, role, is_active, default_addr)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;

-- name: UpdateUser :one
UPDATE users
SET full_name    = $2,
    phone        = $3,
    default_addr = $4,
    updated_at   = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateUserActive :one
UPDATE users
SET is_active  = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;
