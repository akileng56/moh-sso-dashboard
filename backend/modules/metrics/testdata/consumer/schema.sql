-- Disposable test schema only; production tables remain owned by platform migrations.
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
CREATE TABLE client (
    id              UUID PRIMARY KEY,           
    client_id       TEXT NOT NULL UNIQUE,
    name            TEXT NOT NULL,
    description     TEXT,
    icon            TEXT,
    base_url        TEXT,
    root_url        TEXT,
    admin_url       TEXT,
    public_client   BOOLEAN NOT NULL DEFAULT false,
    redirect_uris   TEXT[] NOT NULL DEFAULT '{}',
    web_origins     TEXT[] NOT NULL DEFAULT '{}',
    enabled         BOOLEAN NOT NULL DEFAULT true,
    attributes      JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE audit_logs (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID REFERENCES users(id) ON DELETE SET NULL,
    action     VARCHAR(200) NOT NULL,
    metadata   JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

INSERT INTO users(id,username,email,enabled,roles)
VALUES('00000000-0000-0000-0000-000000000001','one','one@example.org',true,ARRAY['user']),
('00000000-0000-0000-0000-000000000002','two','two@example.org',false,ARRAY['user']),
('00000000-0000-0000-0000-000000000003','three','three@example.org',true,ARRAY['user']);
INSERT INTO client(id,client_id,name,enabled,public_client)
VALUES('10000000-0000-0000-0000-000000000001','client-one','One',true,false),
('10000000-0000-0000-0000-000000000002','client-two','Two',false,true);
INSERT INTO audit_logs(user_id,action,metadata)
VALUES('00000000-0000-0000-0000-000000000001','login','{"success":true,"client_id":"client-one","country":"UG"}'),
('00000000-0000-0000-0000-000000000002','login','{"success":false,"client_id":"client-two","country":"UG"}'),
(NULL,'login','{"success":true,"country":"KE"}'),
('00000000-0000-0000-0000-000000000001','password_reset','{}');
