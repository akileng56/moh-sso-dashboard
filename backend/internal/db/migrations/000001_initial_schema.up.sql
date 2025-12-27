-- ============================================================
--  Initial Schema for moh-sso-dashboard (Keycloak as Source of Truth)
-- ============================================================

-- Enable UUID extension (already fine)
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================
--  USERS (ID = Keycloak user_id)
-- ============================================================
CREATE TABLE users (
    id          UUID PRIMARY KEY,              -- <-- CHANGED: NO DEFAULT
    username    VARCHAR(100) UNIQUE NOT NULL,
    first_name  VARCHAR(100),
    last_name   VARCHAR(100),
    email       VARCHAR(255) UNIQUE NOT NULL,
    enabled     BOOLEAN DEFAULT TRUE,
    role        VARCHAR(50) NOT NULL,
    created_at  TIMESTAMP DEFAULT NOW(),
    updated_at  TIMESTAMP DEFAULT NOW(),
    last_login_at TIMESTAMP DEFAULT NOW()
);

-- ============================================================
--  ROLES
-- ============================================================
CREATE TABLE roles (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    created_at  TIMESTAMP DEFAULT NOW()
);

-- ============================================================
--  USER ROLES (Many-to-Many)
-- ============================================================
CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

-- ============================================================
--  CLIENTS (Registered Applications)
-- ============================================================
CREATE TABLE client (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id       VARCHAR(100) UNIQUE NOT NULL,
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    base_url        TEXT,
    icon            TEXT,
    public_client   BOOLEAN DEFAULT FALSE,
    enabled         BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMP DEFAULT NOW(),
    updated_at      TIMESTAMP DEFAULT NOW()
);

-- ============================================================
--  CLIENT SECRETS (Multiple Secrets per Client)
-- ============================================================
CREATE TABLE client_secrets (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id   UUID NOT NULL REFERENCES client(id) ON DELETE CASCADE,
    secret_hash TEXT NOT NULL,
    created_at  TIMESTAMP DEFAULT NOW(),
    expires_at  TIMESTAMP
);

-- ============================================================
--  USER ↔ CLIENT ACCESS
-- ============================================================
CREATE TABLE user_client_access (
    user_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    client_id UUID NOT NULL REFERENCES client(id) ON DELETE CASCADE,
    granted_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (user_id, client_id)
);

-- ============================================================
--  LOGIN SESSIONS
-- ============================================================
CREATE TABLE sessions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token   TEXT NOT NULL,
    user_agent      TEXT,
    ip_address      TEXT,
    created_at      TIMESTAMP DEFAULT NOW(),
    expires_at      TIMESTAMP NOT NULL
);

-- ============================================================
--  AUDIT LOGS
-- ============================================================
CREATE TABLE audit_logs (
    id         BIGSERIAL PRIMARY KEY,
    user_id    UUID REFERENCES users(id) ON DELETE SET NULL,
    action     VARCHAR(200) NOT NULL,
    metadata   JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

-- ============================================================
--  REVOKED TOKENS
-- ============================================================
CREATE TABLE revoked_tokens (
    id          BIGSERIAL PRIMARY KEY,
    token_hash  TEXT UNIQUE NOT NULL,
    revoked_at  TIMESTAMP DEFAULT NOW()
);


CREATE TYPE import_job_status AS ENUM ('previewed', 'running', 'completed', 'failed');

CREATE TABLE IF NOT EXISTS import_jobs (
  id            UUID PRIMARY KEY,
  filename      TEXT NOT NULL,
  status        import_job_status NOT NULL DEFAULT 'previewed',
  total_rows    INT  NOT NULL DEFAULT 0,
  valid_rows    INT  NOT NULL DEFAULT 0,
  success_count INT  NOT NULL DEFAULT 0,
  failure_count INT  NOT NULL DEFAULT 0,
  created_by    TEXT NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS import_job_items (
  id          UUID PRIMARY KEY,
  job_id      UUID NOT NULL REFERENCES import_jobs(id) ON DELETE CASCADE,
  row_number  INT  NOT NULL,
  username    TEXT,
  email       TEXT,
  first_name  TEXT,
  last_name   TEXT,
  role        TEXT,
  enabled     BOOLEAN,
  client_ids  TEXT,  -- raw string from CSV, e.g. "app1,app2"
  status      TEXT NOT NULL DEFAULT 'pending', -- pending|success|failed|skipped
  error_msg   TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_import_job_items_job ON import_job_items(job_id);