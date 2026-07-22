-- name: ListRecordingsByCameraRange :many
SELECT id, camera_id, started_at, duration_sec, size_bytes, path, codec,
       thumbnail_path, archived_at, archive_location, created_at
FROM recordings
WHERE camera_id = sqlc.arg(camera_id)
  AND started_at >= sqlc.arg(from_ts)
  AND started_at <= sqlc.arg(to_ts)
ORDER BY started_at ASC;

-- name: GetRecording :one
SELECT id, camera_id, started_at, duration_sec, size_bytes, path, codec,
       thumbnail_path, archived_at, archive_location, created_at
FROM recordings
WHERE id = ?;

-- name: InsertRecording :one
INSERT INTO recordings (
  id, camera_id, started_at, duration_sec, size_bytes, path, codec, thumbnail_path
) VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?
)
RETURNING id, camera_id, started_at, duration_sec, size_bytes, path, codec,
          thumbnail_path, archived_at, archive_location, created_at;

-- name: DeleteRecording :exec
DELETE FROM recordings WHERE id = ?;

-- name: DeleteRecordingsOlderThan :execrows
DELETE FROM recordings WHERE started_at < sqlc.arg(before_ts);
