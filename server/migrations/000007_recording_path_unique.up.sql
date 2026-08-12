DELETE FROM recordings
WHERE rowid NOT IN (
  SELECT MIN(rowid)
  FROM recordings
  GROUP BY path
);

CREATE UNIQUE INDEX idx_recordings_path_unique ON recordings(path);
