-- name: ListCameras :many
SELECT id, name, driver, host, username, password_enc, enabled, status, created_at, updated_at
FROM cameras
ORDER BY name ASC, created_at ASC;

-- name: GetCamera :one
SELECT id, name, driver, host, username, password_enc, enabled, status, created_at, updated_at
FROM cameras
WHERE id = ?;

-- name: CreateCamera :one
INSERT INTO cameras (id, name, driver, host, username, password_enc, enabled, status)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING id, name, driver, host, username, password_enc, enabled, status, created_at, updated_at;

-- name: UpdateCamera :one
UPDATE cameras
SET name = ?,
    driver = ?,
    host = ?,
    username = ?,
    password_enc = ?,
    enabled = ?,
    status = ?,
    updated_at = datetime('now')
WHERE id = ?
RETURNING id, name, driver, host, username, password_enc, enabled, status, created_at, updated_at;

-- name: DeleteCamera :exec
DELETE FROM cameras WHERE id = ?;

-- name: UpdateCameraStatus :exec
UPDATE cameras
SET status = ?,
    updated_at = datetime('now')
WHERE id = ?;
