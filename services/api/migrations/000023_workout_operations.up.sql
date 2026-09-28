-- Retain receipts for the lifetime of the account: offline queues can retry
-- after an unbounded absence. Account deletion cascades through this table.
CREATE TABLE workout_operations (
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  operation_id TEXT NOT NULL CHECK (length(operation_id) BETWEEN 8 AND 128),
  workout_id UUID NOT NULL,
  kind TEXT NOT NULL CHECK (kind IN ('log_set','finish_workout','cancel_workout')),
  payload_hash TEXT NOT NULL,
  result JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, operation_id)
);
