-- name: ListStreamProfilesByCamera :many
SELECT id, camera_id, role, rtsp_url, codec, width, height
FROM stream_profiles
WHERE camera_id = ?
ORDER BY role ASC;

-- name: ListStreamProfilesByCameraIDs :many
SELECT id, camera_id, role, rtsp_url, codec, width, height
FROM stream_profiles
WHERE camera_id IN (sqlc.slice('camera_ids'))
ORDER BY camera_id ASC, role ASC;

-- name: UpsertStreamProfile :one
INSERT INTO stream_profiles (id, camera_id, role, rtsp_url, codec, width, height)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(camera_id, role) DO UPDATE SET
  rtsp_url = excluded.rtsp_url,
  codec = excluded.codec,
  width = excluded.width,
  height = excluded.height
RETURNING id, camera_id, role, rtsp_url, codec, width, height;

-- name: DeleteStreamProfilesByCamera :exec
DELETE FROM stream_profiles WHERE camera_id = ?;
