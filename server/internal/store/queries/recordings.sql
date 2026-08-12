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

-- name: GetRecordingByPath :one
SELECT id, camera_id, started_at, duration_sec, size_bytes, path, codec,
       thumbnail_path, archived_at, archive_location, created_at
FROM recordings
WHERE path = ?;

-- name: InsertRecording :one
INSERT INTO recordings (
  id, camera_id, started_at, duration_sec, size_bytes, path, codec, thumbnail_path
) VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?
)
ON CONFLICT(path) DO UPDATE SET
  camera_id = excluded.camera_id,
  started_at = excluded.started_at,
  duration_sec = excluded.duration_sec,
  size_bytes = excluded.size_bytes,
  codec = COALESCE(excluded.codec, recordings.codec),
  thumbnail_path = COALESCE(excluded.thumbnail_path, recordings.thumbnail_path)
RETURNING id, camera_id, started_at, duration_sec, size_bytes, path, codec,
          thumbnail_path, archived_at, archive_location, created_at;

-- name: DeleteRecording :exec
DELETE FROM recordings WHERE id = ?;

-- name: DeleteRecordingsOlderThan :execrows
DELETE FROM recordings WHERE started_at < sqlc.arg(before_ts);

-- name: DeleteArchivedRecordingsOlderThan :execrows
DELETE FROM recordings
WHERE started_at < sqlc.arg(before_ts)
  AND archived_at IS NOT NULL
  AND archived_at != '';

-- name: ListUnarchivedRecordings :many
SELECT id, camera_id, started_at, duration_sec, size_bytes, path, codec,
       thumbnail_path, archived_at, archive_location, created_at
FROM recordings
WHERE archived_at IS NULL OR archived_at = ''
ORDER BY started_at ASC
LIMIT ?;

-- name: MarkRecordingArchived :execrows
UPDATE recordings
SET archived_at = ?, archive_location = ?
WHERE id = ?;
