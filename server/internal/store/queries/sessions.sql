-- name: CreateSession :one
INSERT INTO sessions (id, user_id, token_hash, expires_at)
VALUES (?, ?, ?, ?)
RETURNING id, user_id, token_hash, expires_at, created_at;

-- name: GetSessionByTokenHash :one
SELECT
  s.id AS session_id,
  s.user_id,
  s.token_hash,
  s.expires_at,
  s.created_at AS session_created_at,
  u.username,
  u.role
FROM sessions s
INNER JOIN users u ON u.id = s.user_id
WHERE s.token_hash = ?;

-- name: DeleteSessionByTokenHash :exec
DELETE FROM sessions WHERE token_hash = ?;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions WHERE expires_at < datetime('now');
