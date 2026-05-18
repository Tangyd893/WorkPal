CREATE TABLE IF NOT EXISTS approval_templates (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    form_schema JSONB NOT NULL DEFAULT '{}',
    flow_definition JSONB NOT NULL DEFAULT '{}',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS approval_instances (
    id BIGSERIAL PRIMARY KEY,
    template_id BIGINT NOT NULL REFERENCES approval_templates(id),
    project_id BIGINT,
    title VARCHAR(500) NOT NULL,
    form_data JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(30) NOT NULL DEFAULT 'pending',
    submitter_id BIGINT NOT NULL,
    current_node_id VARCHAR(64),
    submitted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_approval_instances_submitter ON approval_instances(submitter_id);
CREATE INDEX IF NOT EXISTS idx_approval_instances_status ON approval_instances(status);

CREATE TABLE IF NOT EXISTS approval_actions (
    id BIGSERIAL PRIMARY KEY,
    instance_id BIGINT NOT NULL REFERENCES approval_instances(id) ON DELETE CASCADE,
    node_id VARCHAR(64) NOT NULL,
    action VARCHAR(20) NOT NULL,
    comment TEXT DEFAULT '',
    user_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_approval_actions_instance ON approval_actions(instance_id);
