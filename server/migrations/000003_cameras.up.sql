CREATE TABLE cameras (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  driver TEXT NOT NULL DEFAULT 'generic-rtsp',
  host TEXT NOT NULL DEFAULT '',
  username TEXT NOT NULL DEFAULT '',
  password_enc TEXT NOT NULL DEFAULT '',
  enabled INTEGER NOT NULL DEFAULT 1,
  status TEXT NOT NULL DEFAULT 'unknown' CHECK (status IN ('online','offline','unknown')),
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE stream_profiles (
  id TEXT PRIMARY KEY,
  camera_id TEXT NOT NULL REFERENCES cameras(id) ON DELETE CASCADE,
  role TEXT NOT NULL CHECK (role IN ('live','record')),
  rtsp_url TEXT NOT NULL,
  codec TEXT,
  width INTEGER,
  height INTEGER,
  UNIQUE(camera_id, role)
);

CREATE INDEX idx_stream_profiles_camera ON stream_profiles(camera_id);
