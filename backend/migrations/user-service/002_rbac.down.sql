-- +migrate Down
DROP TABLE IF EXISTS project_members;
DROP TABLE IF EXISTS project_role_permissions;
DROP TABLE IF EXISTS project_roles;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
