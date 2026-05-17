-- IM服务增强：频道和话题线程 (回滚)
DROP INDEX IF EXISTS idx_messages_message_type;
DROP INDEX IF EXISTS idx_messages_thread_id;
ALTER TABLE messages
    DROP COLUMN IF EXISTS metadata,
    DROP COLUMN IF EXISTS message_type,
    DROP COLUMN IF EXISTS thread_id;
DROP TABLE IF EXISTS threads;
DROP TABLE IF EXISTS channel_members;
DROP TABLE IF EXISTS channels;
