-- name: CreateUser :one
INSERT INTO users (email, username, password_hash)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmailForAuth :one
SELECT * FROM users WHERE lower(email) = lower($1);

-- name: ExistsUserByEmail :one
SELECT EXISTS (SELECT 1 FROM users WHERE lower(email) = lower($1));