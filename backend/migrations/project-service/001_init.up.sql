-- 项目管理服务初始化迁移
-- +migrate Up
CREATE TABLE IF NOT EXISTS projects (
    id BIGSERIAL PRIMARY KEY,
    key VARCHAR(10) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    lead_id BIGINT NOT NULL,
    icon VARCHAR(50) DEFAULT '',
    category VARCHAR(50) DEFAULT 'software',
    is_archived BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX IF NOT EXISTS idx_projects_key ON projects(key);
CREATE INDEX IF NOT EXISTS idx_projects_deleted_at ON projects(deleted_at);

CREATE TABLE IF NOT EXISTS issue_types (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT DEFAULT '',
    icon_url VARCHAR(255) DEFAULT '',
    hierarchy_level INT DEFAULT 0,
    is_standard BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_issue_types_project_id ON issue_types(project_id);

CREATE TABLE IF NOT EXISTS workflows (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    dsl_definition JSONB NOT NULL DEFAULT '{}',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_workflows_project_id ON workflows(project_id);

CREATE TABLE IF NOT EXISTS issues (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    issue_type_id BIGINT NOT NULL REFERENCES issue_types(id),
    parent_id BIGINT REFERENCES issues(id),
    key VARCHAR(50) NOT NULL UNIQUE,
    summary VARCHAR(500) NOT NULL,
    description TEXT DEFAULT '',
    status VARCHAR(50) NOT NULL DEFAULT 'Open',
    priority VARCHAR(20) NOT NULL DEFAULT 'Medium',
    assignee_id BIGINT,
    reporter_id BIGINT NOT NULL,
    due_date DATE,
    story_points DECIMAL(5,1),
    resolution VARCHAR(50) DEFAULT '',
    sprint_id BIGINT,
    version_ids BIGINT[] DEFAULT '{}',
    fix_version_ids BIGINT[] DEFAULT '{}',
    time_estimate INT DEFAULT 0,
    time_spent INT DEFAULT 0,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX IF NOT EXISTS idx_issues_project_id ON issues(project_id);
CREATE INDEX IF NOT EXISTS idx_issues_key ON issues(key);
CREATE INDEX IF NOT EXISTS idx_issues_assignee_id ON issues(assignee_id);
CREATE INDEX IF NOT EXISTS idx_issues_status ON issues(status);
CREATE INDEX IF NOT EXISTS idx_issues_parent_id ON issues(parent_id);
CREATE INDEX IF NOT EXISTS idx_issues_deleted_at ON issues(deleted_at);

CREATE TABLE IF NOT EXISTS boards (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    board_type VARCHAR(20) NOT NULL DEFAULT 'kanban',
    config JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_boards_project_id ON boards(project_id);

CREATE TABLE IF NOT EXISTS versions (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    start_date DATE,
    release_date DATE,
    is_archived BOOLEAN DEFAULT FALSE,
    is_released BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_versions_project_id ON versions(project_id);

CREATE TABLE IF NOT EXISTS issue_changelogs (
    id BIGSERIAL PRIMARY KEY,
    issue_id BIGINT NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    field VARCHAR(100) NOT NULL,
    old_value TEXT DEFAULT '',
    new_value TEXT DEFAULT '',
    changed_by BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_issue_changelogs_issue_id ON issue_changelogs(issue_id);

CREATE TABLE IF NOT EXISTS custom_field_defs (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    field_type VARCHAR(30) NOT NULL DEFAULT 'text',
    options JSONB DEFAULT '[]',
    is_required BOOLEAN DEFAULT FALSE,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_custom_field_defs_project_id ON custom_field_defs(project_id);

CREATE TABLE IF NOT EXISTS custom_field_values (
    id BIGSERIAL PRIMARY KEY,
    issue_id BIGINT NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    field_id BIGINT NOT NULL REFERENCES custom_field_defs(id) ON DELETE CASCADE,
    value_text TEXT DEFAULT '',
    value_number DECIMAL DEFAULT 0,
    value_date DATE,
    value_json JSONB DEFAULT '{}',
    UNIQUE(issue_id, field_id)
);
CREATE INDEX IF NOT EXISTS idx_custom_field_values_issue_id ON custom_field_values(issue_id);

CREATE TABLE IF NOT EXISTS associations (
    id BIGSERIAL PRIMARY KEY,
    source_type VARCHAR(50) NOT NULL,
    source_id BIGINT NOT NULL,
    target_type VARCHAR(50) NOT NULL,
    target_id BIGINT NOT NULL,
    link_type VARCHAR(50) NOT NULL DEFAULT 'related',
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_associations_source ON associations(source_type, source_id);
CREATE INDEX IF NOT EXISTS idx_associations_target ON associations(target_type, target_id);
