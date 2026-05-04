-- 文件服务初始化迁移
-- +migrate Up
CREATE TABLE IF NOT EXISTS files (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    conv_id BIGINT DEFAULT 0,
    name VARCHAR(256) NOT NULL,
    key VARCHAR(512) NOT NULL,
    size BIGINT NOT NULL,
    content_type VARCHAR(128) DEFAULT '',
    mime_type VARCHAR(128) DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_files_user_id ON files(user_id);
CREATE INDEX IF NOT EXISTS idx_files_conv_id ON files(conv_id);

-- +migrate Down
DROP TABLE IF EXISTS files;
