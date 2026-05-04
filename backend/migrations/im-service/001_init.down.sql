-- IM服务初始化迁移 (回滚)
DROP TABLE IF EXISTS message_reads;
DROP TABLE IF EXISTS message_outbox;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS conversation_members;
DROP TABLE IF EXISTS conversations;
