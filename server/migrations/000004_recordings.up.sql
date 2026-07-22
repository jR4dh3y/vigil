CREATE TABLE recordings (
  id TEXT PRIMARY KEY,
  camera_id TEXT NOT NULL REFERENCES cameras(id) ON DELETE CASCADE,
  started_at TEXT NOT NULL,
  duration_sec REAL NOT NULL,
  size_bytes INTEGER NOT NULL DEFAULT 0,
  path TEXT NOT NULL,
  codec TEXT,
  thumbnail_path TEXT,
  archived_at TEXT,
  archive_location TEXT,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_recordings_camera_started ON recordings(camera_id, started_at);
