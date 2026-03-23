-- name: CreateUser :one
INSERT INTO users (username, email, password)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserLevel :one
SELECT level FROM users WHERE id = $1;

-- name: ListUsers :many
SELECT id, username, email, level, created_at FROM users ORDER BY id;

-- name: UpdateUserLevel :exec
UPDATE users SET level = $2, updated_at = NOW()
WHERE id = $1;
