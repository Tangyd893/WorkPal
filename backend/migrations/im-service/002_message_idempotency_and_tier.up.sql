ALTER TABLE messages
  ADD COLUMN IF NOT EXISTS idempotency_key VARCHAR(128) NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS tier VARCHAR(16) NOT NULL DEFAULT 'hot';

CREATE UNIQUE INDEX IF NOT EXISTS uniq_messages_idempotency_active
  ON messages (conv_id, sender_id, idempotency_key)
  WHERE idempotency_key <> '';

CREATE INDEX IF NOT EXISTS idx_messages_conv_created_at_active
  ON messages (conv_id, created_at DESC)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_messages_conv_tier_created_at_active
  ON messages (conv_id, tier, created_at DESC)
  WHERE deleted_at IS NULL;
