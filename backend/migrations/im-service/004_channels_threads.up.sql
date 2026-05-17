-- IM服务增强：频道和话题线程
-- +migrate Up
CREATE TABLE IF NOT EXISTS channels (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    channel_type VARCHAR(20) NOT NULL DEFAULT 'public',
    created_by BIGINT NOT NULL,
    is_archived BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_channels_project_id ON channels(project_id);
CREATE INDEX IF NOT EXISTS idx_channels_channel_type ON channels(channel_type);

CREATE TABLE IF NOT EXISTS channel_members (
    id BIGSERIAL PRIMARY KEY,
    channel_id BIGINT NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'member',
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(channel_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_channel_members_channel_id ON channel_members(channel_id);
CREATE INDEX IF NOT EXISTS idx_channel_members_user_id ON channel_members(user_id);

CREATE TABLE IF NOT EXISTS threads (
    id BIGSERIAL PRIMARY KEY,
    channel_id BIGINT NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    parent_msg_id BIGINT NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    title VARCHAR(500) DEFAULT '',
    reply_count INT DEFAULT 0,
    last_reply_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_threads_channel_id ON threads(channel_id);
CREATE INDEX IF NOT EXISTS idx_threads_parent_msg_id ON threads(parent_msg_id);

ALTER TABLE messages
    ADD COLUMN IF NOT EXISTS thread_id BIGINT REFERENCES threads(id),
    ADD COLUMN IF NOT EXISTS message_type VARCHAR(30) DEFAULT 'text',
    ADD COLUMN IF NOT EXISTS metadata JSONB DEFAULT '{}';
CREATE INDEX IF NOT EXISTS idx_messages_thread_id ON messages(thread_id);
CREATE INDEX IF NOT EXISTS idx_messages_message_type ON messages(message_type);
