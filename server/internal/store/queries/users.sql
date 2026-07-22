-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: CreateUser :one
INSERT INTO users (id, username, password_hash, role)
VALUES (?, ?, ?, ?)
RETURNING id, username, password_hash, role, created_at, updated_at;

-- name: GetUserByUsername :one
SELECT id, username, password_hash, role, created_at, updated_at
FROM users
WHERE username = ?;

-- name: GetUserByID :one
SELECT id, username, password_hash, role, created_at, updated_at
FROM users
WHERE id = ?;

-- name: ListUsers :many
SELECT id, username, password_hash, role, created_at, updated_at
FROM users
ORDER BY username ASC;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;

-- name: CountAdmins :one
SELECT COUNT(*) FROM users WHERE role = 'admin';
