-- 用户服务初始化迁移 (回滚)
-- +migrate Down
DROP TABLE IF EXISTS employees;
DROP TABLE IF EXISTS departments;
DROP TABLE IF EXISTS users;
