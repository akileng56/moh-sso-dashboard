-- Disposable fixture schema; primary production migrations remain platform-owned.
CREATE TABLE users (
    id            UUID PRIMARY KEY,
    username      VARCHAR(100) UNIQUE NOT NULL,
    first_name    VARCHAR(100),
    last_name     VARCHAR(100),
    email         VARCHAR(255) UNIQUE NOT NULL,
    enabled       BOOLEAN DEFAULT TRUE,
    roles         TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW(),
    last_login_at TIMESTAMPTZ
);
CREATE TABLE audit_logs (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID REFERENCES users(id) ON DELETE SET NULL,
    action     VARCHAR(200) NOT NULL,
    metadata   JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

INSERT INTO users(id,username,email)
VALUES('00000000-0000-0000-0000-000000000001','fixture','fixture@example.org');
INSERT INTO audit_logs(id,user_id,action,metadata,created_at)
VALUES('10000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000001','login','{"success":true,"client_id":"fixture"}','2026-11-19T12:00:00Z'),
('10000000-0000-0000-0000-000000000002','00000000-0000-0000-0000-000000000001','login','{"success":false,"ip":"127.0.0.1"}','2026-11-19T12:00:00Z'),
('10000000-0000-0000-0000-000000000003',NULL,'system-event',NULL,'2026-11-19T12:00:00Z');
