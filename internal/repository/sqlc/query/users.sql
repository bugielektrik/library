-- name: CreateUser :one
INSERT INTO users (id, email, password_hash, full_name, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, full_name, created_at, updated_at
FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT id, email, password_hash, full_name, created_at, updated_at
FROM users
WHERE id = $1;