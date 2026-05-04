-- IM服务初始化迁移
-- +migrate Up
CREATE TABLE IF NOT EXISTS conversations (
    id BIGSERIAL PRIMARY KEY,
    type SMALLINT DEFAULT 1 NOT NULL,
    name VARCHAR(256) DEFAULT '',
    avatar_url VARCHAR(512) DEFAULT '',
    owner_id BIGINT DEFAULT 0,
    announcement TEXT DEFAULT '',
    announcement_updated_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX IF NOT EXISTS idx_conversations_deleted_at ON conversations(deleted_at);

CREATE TABLE IF NOT EXISTS conversation_members (
    id BIGSERIAL PRIMARY KEY,
    conv_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (conv_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_conversation_members_conv_id ON conversation_members(conv_id);
CREATE INDEX IF NOT EXISTS idx_conversation_members_user_id ON conversation_members(user_id);

CREATE TABLE IF NOT EXISTS messages (
    id BIGSERIAL PRIMARY KEY,
    conv_id BIGINT NOT NULL,
    sender_id BIGINT NOT NULL,
    type SMALLINT DEFAULT 1 NOT NULL,
    content TEXT DEFAULT '',
    metadata JSONB DEFAULT '{}',
    reply_to BIGINT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX IF NOT EXISTS idx_messages_conv_id ON messages(conv_id);
CREATE INDEX IF NOT EXISTS idx_messages_sender_id ON messages(sender_id);
CREATE INDEX IF NOT EXISTS idx_messages_deleted_at ON messages(deleted_at);

CREATE TABLE IF NOT EXISTS message_outbox (
    id BIGSERIAL PRIMARY KEY,
    topic VARCHAR(128) DEFAULT '',
    payload JSONB DEFAULT '{}',
    status VARCHAR(32) DEFAULT 'pending',
    retry_count INT DEFAULT 0 NOT NULL,
    last_error TEXT DEFAULT '',
    next_attempt_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    delivered_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_message_outbox_topic ON message_outbox(topic);
CREATE INDEX IF NOT EXISTS idx_message_outbox_status ON message_outbox(status);
CREATE INDEX IF NOT EXISTS idx_message_outbox_next_attempt_at ON message_outbox(next_attempt_at);

CREATE TABLE IF NOT EXISTS message_reads (
    user_id BIGINT NOT NULL,
    conv_id BIGINT NOT NULL,
    read_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (user_id, conv_id)
);

-- +migrate Down
DROP TABLE IF EXISTS message_reads;
DROP TABLE IF EXISTS message_outbox;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS conversation_members;
DROP TABLE IF EXISTS conversations;
