-- name: InsertEvent :one
INSERT INTO events (
  id, camera_id, type, severity, title, message,
  started_at, ended_at, metadata, thumbnail_path, acknowledged
) VALUES (
  ?, ?, ?, ?, ?, ?,
  ?, ?, ?, ?, ?
)
RETURNING id, camera_id, type, severity, title, message,
  started_at, ended_at, metadata, thumbnail_path, acknowledged, created_at;

-- name: GetEvent :one
SELECT id, camera_id, type, severity, title, message,
  started_at, ended_at, metadata, thumbnail_path, acknowledged, created_at
FROM events
WHERE id = ?;

-- name: ListEvents :many
SELECT id, camera_id, type, severity, title, message,
  started_at, ended_at, metadata, thumbnail_path, acknowledged, created_at
FROM events
WHERE (cast(sqlc.narg('camera_id') AS TEXT) IS NULL OR camera_id = sqlc.narg('camera_id'))
  AND (cast(sqlc.narg('event_type') AS TEXT) IS NULL OR type = sqlc.narg('event_type'))
  AND (cast(sqlc.narg('before') AS TEXT) IS NULL OR started_at < sqlc.narg('before'))
  AND (
    cast(sqlc.narg('unacked_only') AS INTEGER) IS NULL
    OR sqlc.narg('unacked_only') = 0
    OR acknowledged = 0
  )
ORDER BY started_at DESC, id DESC
LIMIT sqlc.arg('limit_count');

-- name: AcknowledgeEvent :one
UPDATE events
SET acknowledged = 1
WHERE id = ?
RETURNING id, camera_id, type, severity, title, message,
  started_at, ended_at, metadata, thumbnail_path, acknowledged, created_at;
