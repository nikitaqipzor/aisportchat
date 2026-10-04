-- A key is unique per capture. Keep cleanup independent of scans and users so
-- account and scan removal cannot discard a failed file deletion.
CREATE TABLE body_scan_media_cleanup (
  storage_key TEXT PRIMARY KEY,
  user_id UUID NOT NULL,
  ready_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX body_scan_media_cleanup_ready ON body_scan_media_cleanup(ready_at);
