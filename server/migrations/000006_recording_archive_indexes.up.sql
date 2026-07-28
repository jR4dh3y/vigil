CREATE INDEX idx_recordings_unarchived_started
ON recordings(started_at)
WHERE archived_at IS NULL OR archived_at = '';

CREATE INDEX idx_recordings_archived_started
ON recordings(started_at)
WHERE archived_at IS NOT NULL AND archived_at != '';
