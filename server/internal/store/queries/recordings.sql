-- name: ListRecordingsByCameraRange :many
SELECT id, camera_id, started_at, duration_sec, size_bytes, path, codec,
       thumbnail_path, archived_at, archive_location, created_at
FROM recordings
WHERE camera_id = sqlc.arg(camera_id)
  AND started_at >= sqlc.arg(from_ts)
  AND started_at <= sqlc.arg(to_ts)
ORDER BY started_at ASC;

-- name: ListRecordingStorageByCameraRange :many
SELECT started_at, duration_sec, archive_location
FROM recordings
WHERE camera_id = sqlc.arg(camera_id)
  AND started_at <= sqlc.arg(to_ts)
  AND julianday(started_at) + duration_sec / 86400.0 >= julianday(sqlc.arg(from_ts))
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

-- name: GetRecordingAtOrBefore :one
SELECT id, camera_id, started_at, duration_sec, size_bytes, path, codec,
       thumbnail_path, archived_at, archive_location, created_at
FROM recordings
WHERE camera_id = sqlc.arg(camera_id)
  AND started_at <= sqlc.arg(at_ts)
ORDER BY started_at DESC
LIMIT 1;

-- name: GetNextRecording :one
SELECT id, camera_id, started_at, duration_sec, size_bytes, path, codec,
       thumbnail_path, archived_at, archive_location, created_at
FROM recordings
WHERE camera_id = sqlc.arg(camera_id)
  AND started_at > sqlc.arg(after_ts)
ORDER BY started_at ASC
LIMIT 1;

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
DELETE FROM recordings
WHERE started_at < sqlc.arg(before_ts)
  AND (archive_location IS NULL OR archive_location = '' OR archive_location NOT LIKE 'gdrive:%');

-- name: DeleteArchivedRecordingsOlderThan :execrows
DELETE FROM recordings
WHERE started_at < sqlc.arg(before_ts)
  AND archived_at IS NOT NULL
  AND archived_at != ''
  AND (archive_location IS NULL OR archive_location = '' OR archive_location NOT LIKE 'gdrive:%');

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
