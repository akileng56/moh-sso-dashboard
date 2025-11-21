-- ============================================================
--  Rollback Initial Schema for moh-sso-dashboard
-- ============================================================

DROP TABLE IF EXISTS revoked_tokens;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS user_client_access;
DROP TABLE IF EXISTS client_secrets;
DROP TABLE IF EXISTS client;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS user;

DROP EXTENSION IF EXISTS "pgcrypto";
