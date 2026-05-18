-- RBAC 权限模型迁移
-- +migrate Up
CREATE TABLE IF NOT EXISTS roles (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(64) NOT NULL UNIQUE,
    description VARCHAR(255) DEFAULT '',
    is_system BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS permissions (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(128) NOT NULL UNIQUE,
    name VARCHAR(128) NOT NULL,
    description VARCHAR(255) DEFAULT '',
    resource_type VARCHAR(64) NOT NULL DEFAULT 'global',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id BIGINT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    project_id BIGINT DEFAULT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id, COALESCE(project_id, 0))
);

CREATE TABLE IF NOT EXISTS project_roles (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL,
    name VARCHAR(64) NOT NULL,
    description VARCHAR(255) DEFAULT '',
    is_system BOOLEAN DEFAULT TRUE,
    UNIQUE(project_id, name)
);

CREATE TABLE IF NOT EXISTS project_role_permissions (
    project_role_id BIGINT NOT NULL REFERENCES project_roles(id) ON DELETE CASCADE,
    permission_id BIGINT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (project_role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS project_members (
    user_id BIGINT NOT NULL,
    project_id BIGINT NOT NULL,
    project_role_id BIGINT NOT NULL REFERENCES project_roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (user_id, project_id)
);

-- 预置系统角色
INSERT INTO roles (name, description, is_system) VALUES
    ('系统管理员', '拥有系统全部管理权限', TRUE),
    ('用户管理员', '管理用户和部门', TRUE),
    ('全局浏览者', '只读浏览权限', TRUE)
ON CONFLICT (name) DO NOTHING;

-- 预置权限
INSERT INTO permissions (code, name, resource_type) VALUES
    ('project:create', '创建项目', 'project'),
    ('project:delete', '删除项目', 'project'),
    ('project:edit', '编辑项目', 'project'),
    ('project:view', '查看项目', 'project'),
    ('issue:create', '创建事项', 'issue'),
    ('issue:edit', '编辑事项', 'issue'),
    ('issue:delete', '删除事项', 'issue'),
    ('issue:view', '查看事项', 'issue'),
    ('issue:change_status', '变更事项状态', 'issue'),
    ('workflow:manage', '管理工作流', 'workflow'),
    ('board:manage', '管理看板', 'board'),
    ('member:manage', '管理项目成员', 'member'),
    ('user:manage', '管理用户', 'user'),
    ('role:manage', '管理角色', 'role')
ON CONFLICT (code) DO NOTHING;

-- 系统管理员 => 全部权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = '系统管理员'
ON CONFLICT DO NOTHING;

-- 用户管理员 => user:manage + role:manage
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = '用户管理员' AND p.code IN ('user:manage', 'role:manage')
ON CONFLICT DO NOTHING;

-- 全局浏览者 => 所有 view 权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = '全局浏览者' AND p.code LIKE '%:view'
ON CONFLICT DO NOTHING;

-- +migrate Down
DROP TABLE IF EXISTS project_members;
DROP TABLE IF EXISTS project_role_permissions;
DROP TABLE IF EXISTS project_roles;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
