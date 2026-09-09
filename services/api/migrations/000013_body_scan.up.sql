CREATE TABLE IF NOT EXISTS body_scans (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','completed')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  completed_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_body_scans_user_created ON body_scans(user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS body_scan_photos (
  id UUID PRIMARY KEY,
  scan_id UUID NOT NULL REFERENCES body_scans(id) ON DELETE CASCADE,
  view TEXT NOT NULL CHECK (view IN ('front','side','back')),
  storage_key TEXT NOT NULL,
  mime_type TEXT NOT NULL,
  width INTEGER NOT NULL,
  height INTEGER NOT NULL,
  bytes INTEGER NOT NULL,
  brightness NUMERIC(6,2) NOT NULL DEFAULT 0,
  contrast NUMERIC(6,2) NOT NULL DEFAULT 0,
  quality_status TEXT NOT NULL CHECK (quality_status IN ('accepted','warning','rejected')),
  quality_issues TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(scan_id, view)
);
