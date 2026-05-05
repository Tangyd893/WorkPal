DROP INDEX IF EXISTS idx_messages_conv_tier_created_at_active;
DROP INDEX IF EXISTS idx_messages_conv_created_at_active;
DROP INDEX IF EXISTS uniq_messages_idempotency_active;

ALTER TABLE messages
  DROP COLUMN IF EXISTS tier,
  DROP COLUMN IF EXISTS idempotency_key;
