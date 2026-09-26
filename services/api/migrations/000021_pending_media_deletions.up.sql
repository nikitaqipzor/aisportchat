-- No FK: the deletion job must survive removal of the owning account.
CREATE TABLE pending_media_deletions (
  user_id UUID PRIMARY KEY,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
