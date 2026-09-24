-- Unique per user and user action. The marker and all food_entries are written
-- in the same transaction, so a failed batch never reserves a retry key.
CREATE TABLE food_entry_operations (
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  operation_key TEXT NOT NULL CHECK (length(operation_key) BETWEEN 8 AND 128),
  payload_hash TEXT NOT NULL,
  logged_at TIMESTAMPTZ NOT NULL,
  PRIMARY KEY (user_id, operation_key)
);
