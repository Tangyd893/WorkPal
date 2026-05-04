-- 工作台服务初始化迁移
-- +migrate Up
CREATE TABLE IF NOT EXISTS tasks (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    summary TEXT DEFAULT '',
    project VARCHAR(128) DEFAULT '',
    owner_username VARCHAR(64) DEFAULT '',
    teammates JSONB DEFAULT '[]',
    due_date VARCHAR(32) DEFAULT '',
    priority VARCHAR(32) DEFAULT 'medium',
    status VARCHAR(32) DEFAULT 'planned',
    shared_count INT DEFAULT 0,
    source VARCHAR(32) DEFAULT 'custom',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON tasks(user_id);
CREATE INDEX IF NOT EXISTS idx_tasks_owner_username ON tasks(owner_username);
CREATE INDEX IF NOT EXISTS idx_tasks_deleted_at ON tasks(deleted_at);

CREATE TABLE IF NOT EXISTS schedule_events (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    detail TEXT DEFAULT '',
    owner_username VARCHAR(64) DEFAULT '',
    starts_at TIMESTAMP WITH TIME ZONE NOT NULL,
    duration_minutes INT DEFAULT 30,
    attendees JSONB DEFAULT '[]',
    room VARCHAR(128) DEFAULT '',
    shared_count INT DEFAULT 0,
    source VARCHAR(32) DEFAULT 'custom',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX IF NOT EXISTS idx_schedule_events_user_id ON schedule_events(user_id);
CREATE INDEX IF NOT EXISTS idx_schedule_events_owner_username ON schedule_events(owner_username);
CREATE INDEX IF NOT EXISTS idx_schedule_events_starts_at ON schedule_events(starts_at);
CREATE INDEX IF NOT EXISTS idx_schedule_events_deleted_at ON schedule_events(deleted_at);

-- +migrate Down
DROP TABLE IF EXISTS schedule_events;
DROP TABLE IF EXISTS tasks;
