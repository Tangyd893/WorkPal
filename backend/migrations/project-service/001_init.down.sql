-- 项目管理服务初始化迁移 (回滚)
DROP TABLE IF EXISTS associations;
DROP TABLE IF EXISTS custom_field_values;
DROP TABLE IF EXISTS custom_field_defs;
DROP TABLE IF EXISTS issue_changelogs;
DROP TABLE IF EXISTS versions;
DROP TABLE IF EXISTS boards;
DROP TABLE IF EXISTS issues;
DROP TABLE IF EXISTS workflows;
DROP TABLE IF EXISTS issue_types;
DROP TABLE IF EXISTS projects;
