-- Test-only snapshot of the host Announcements schema; never applied by NewRepository.
-- Platform tables below include only columns needed by these persistence checks.
CREATE TABLE users(id uuid PRIMARY KEY);
CREATE TABLE ihp_rbac_groups(id uuid PRIMARY KEY);
CREATE TABLE health_context_nodes(id uuid PRIMARY KEY);
CREATE TABLE notification_preferences(user_id uuid PRIMARY KEY, phone_number text,
 sms_enabled boolean NOT NULL, updated_at timestamptz NOT NULL DEFAULT now());

-- Fixture source: 000006_create_announcements_schema.up.sql
-- 20260309_create_announcements.sql

-- Optional: for gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- -------------------------------------
-- Enums
-- -------------------------------------
-- Top-level DDL keeps PostgreSQL and SQLC schema discovery aligned.
CREATE TYPE announcement_status AS ENUM (
    'DRAFT',
    'SCHEDULED',
    'PUBLISHED',
    'ARCHIVED'
);

CREATE TYPE announcement_level AS ENUM (
    'INFO',
    'SUCCESS',
    'WARNING',
    'CRITICAL'
);

CREATE TYPE announcement_audience_type AS ENUM (
    'ALL_USERS',
    'ADMINS_ONLY',
    'SPECIFIC_CLIENTS',
    'SPECIFIC_ROLES',
    'SPECIFIC_USERS'
);

-- -------------------------------------
-- Main announcements table
-- -------------------------------------
CREATE TABLE IF NOT EXISTS announcements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Core content
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    summary VARCHAR(500),
    level announcement_level NOT NULL DEFAULT 'INFO',
    tag VARCHAR(100),
    link_url TEXT,

    -- Display / ordering
    priority INT NOT NULL DEFAULT 0,
    is_pinned BOOLEAN NOT NULL DEFAULT FALSE,

    -- Lifecycle
    status announcement_status NOT NULL DEFAULT 'DRAFT',
    publish_at TIMESTAMP NULL,
    expires_at TIMESTAMP NULL,

    -- Audience
    audience_type announcement_audience_type NOT NULL DEFAULT 'ALL_USERS',

    -- Ownership / audit
    created_by UUID NOT NULL,
    updated_by UUID,
    published_by UUID,
    archived_by UUID,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    published_at TIMESTAMP NULL,
    archived_at TIMESTAMP NULL,

    -- Soft delete
    deleted_at TIMESTAMP NULL,
    deleted_by UUID NULL,

    -- Versioning / concurrency
    version INT NOT NULL DEFAULT 1,

    -- Constraints
    CONSTRAINT chk_announcements_title_not_blank
        CHECK (btrim(title) <> ''),

    CONSTRAINT chk_announcements_message_not_blank
        CHECK (btrim(message) <> ''),

    CONSTRAINT chk_announcements_priority_non_negative
        CHECK (priority >= 0),

    CONSTRAINT chk_announcements_publish_expiry_order
        CHECK (
            expires_at IS NULL
            OR publish_at IS NULL
            OR expires_at > publish_at
        )
);

-- -------------------------------------
-- Audience mapping tables
-- -------------------------------------

-- For client-scoped announcements
CREATE TABLE IF NOT EXISTS announcement_clients (
    announcement_id UUID NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
    client_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    PRIMARY KEY (announcement_id, client_id)
);

-- For role-scoped announcements
CREATE TABLE IF NOT EXISTS announcement_roles (
    announcement_id UUID NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
    role_name VARCHAR(150) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    PRIMARY KEY (announcement_id, role_name)
);

-- For specific-user announcements
CREATE TABLE IF NOT EXISTS announcement_users (
    announcement_id UUID NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    PRIMARY KEY (announcement_id, user_id)
);

-- -------------------------------------
-- Helpful indexes
-- -------------------------------------

-- General sort / feed queries
CREATE INDEX IF NOT EXISTS idx_announcements_created_at
    ON announcements (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_announcements_status
    ON announcements (status);

CREATE INDEX IF NOT EXISTS idx_announcements_level
    ON announcements (level);

CREATE INDEX IF NOT EXISTS idx_announcements_priority_created_at
    ON announcements (priority DESC, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_announcements_publish_at
    ON announcements (publish_at);

CREATE INDEX IF NOT EXISTS idx_announcements_expires_at
    ON announcements (expires_at);

CREATE INDEX IF NOT EXISTS idx_announcements_is_pinned
    ON announcements (is_pinned);

-- Fast active published queries
CREATE INDEX IF NOT EXISTS idx_announcements_active_feed
    ON announcements (is_pinned DESC, priority DESC, publish_at DESC, created_at DESC)
    WHERE deleted_at IS NULL AND status = 'PUBLISHED';

-- Audience mapping indexes
CREATE INDEX IF NOT EXISTS idx_announcement_clients_client_id
    ON announcement_clients (client_id);

CREATE INDEX IF NOT EXISTS idx_announcement_roles_role_name
    ON announcement_roles (role_name);

CREATE INDEX IF NOT EXISTS idx_announcement_users_user_id
    ON announcement_users (user_id);

-- -------------------------------------
-- Trigger to auto-update updated_at + version
-- -------------------------------------
CREATE OR REPLACE FUNCTION set_announcements_updated_fields()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    NEW.version = OLD.version + 1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_announcements_set_updated_fields ON announcements;

CREATE TRIGGER trg_announcements_set_updated_fields
BEFORE UPDATE ON announcements
FOR EACH ROW
EXECUTE FUNCTION set_announcements_updated_fields();

-- Fixture source: 000025_add_announcement_email_notification_columns.up.sql

ALTER TABLE announcements
ADD COLUMN IF NOT EXISTS notify_by_email BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE announcements
ADD COLUMN IF NOT EXISTS email_notification_sent_at TIMESTAMP NULL;

CREATE INDEX IF NOT EXISTS idx_announcements_notify_by_email
    ON announcements (notify_by_email);

CREATE INDEX IF NOT EXISTS idx_announcements_email_notification_pending
    ON announcements (publish_at, created_at)
    WHERE
        deleted_at IS NULL
        AND status = 'PUBLISHED'
        AND notify_by_email = TRUE
        AND email_notification_sent_at IS NULL;

-- Fixture source: 000028_create_announcement_attachments.up.sql
CREATE TABLE IF NOT EXISTS announcement_attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    announcement_id UUID NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
    file_name TEXT NOT NULL,
    original_file_name TEXT NOT NULL,
    content_type TEXT,
    file_size BIGINT NOT NULL,
    storage_provider TEXT NOT NULL,
    storage_key TEXT NOT NULL,
    checksum TEXT,
    uploaded_by UUID REFERENCES users(id) ON DELETE SET NULL,
    include_in_email BOOLEAN NOT NULL DEFAULT TRUE,
    inline BOOLEAN NOT NULL DEFAULT FALSE,
    content_id TEXT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    deleted_by UUID REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_announcement_attachments_announcement_id
    ON announcement_attachments(announcement_id);

CREATE INDEX IF NOT EXISTS idx_announcement_attachments_announcement_deleted
    ON announcement_attachments(announcement_id, deleted_at);

CREATE UNIQUE INDEX IF NOT EXISTS idx_announcement_attachments_storage_key_unique
    ON announcement_attachments(storage_key);


-- Fixture source: 000029_add_announcement_link_label.up.sql
ALTER TABLE announcements
    ADD COLUMN IF NOT EXISTS link_label TEXT;


-- Fixture source: 000036_add_announcement_sms_notification_columns.up.sql
ALTER TABLE announcements
ADD COLUMN IF NOT EXISTS notify_by_sms BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE announcements
ADD COLUMN IF NOT EXISTS sms_message TEXT;

ALTER TABLE announcements
ADD COLUMN IF NOT EXISTS sms_notification_queued_at TIMESTAMPTZ NULL;

CREATE INDEX IF NOT EXISTS idx_announcements_notify_by_sms
    ON announcements (notify_by_sms);

CREATE INDEX IF NOT EXISTS idx_announcements_sms_notification_pending
    ON announcements (publish_at, created_at)
    WHERE
        deleted_at IS NULL
        AND status = 'PUBLISHED'
        AND notify_by_sms = TRUE
        AND sms_notification_queued_at IS NULL;


-- Fixture source: 000039_add_announcement_group_audience.up.sql
ALTER TYPE announcement_audience_type ADD VALUE IF NOT EXISTS 'SPECIFIC_GROUPS';

CREATE TABLE IF NOT EXISTS announcement_groups (
    announcement_id UUID NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
    group_id UUID NOT NULL REFERENCES ihp_rbac_groups(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    PRIMARY KEY (announcement_id, group_id)
);

CREATE INDEX IF NOT EXISTS idx_announcement_groups_group_id
    ON announcement_groups(group_id);


-- Fixture source: 000058_add_announcement_health_context_audience.up.sql
    ALTER TYPE announcement_audience_type
        ADD VALUE IF NOT EXISTS 'SPECIFIC_HEALTH_CONTEXTS';
    ALTER TYPE announcement_audience_type
        ADD VALUE IF NOT EXISTS 'HEALTH_CONTEXT_AND_DESCENDANTS';

CREATE TABLE announcement_health_contexts (
    announcement_id UUID NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
    context_node_id UUID NOT NULL REFERENCES health_context_nodes(id) ON DELETE RESTRICT,
    include_descendants BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (announcement_id, context_node_id)
);

CREATE INDEX announcement_health_contexts_context_idx
    ON announcement_health_contexts(context_node_id);

CREATE TABLE announcement_audience_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    announcement_id UUID NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
    channel TEXT NOT NULL,
    recipient_count INTEGER NOT NULL CHECK (recipient_count >= 0),
    recipient_user_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX announcement_audience_snapshots_announcement_idx
    ON announcement_audience_snapshots(announcement_id, created_at DESC);


INSERT INTO users VALUES ('10000000-0000-0000-0000-000000000001');
INSERT INTO ihp_rbac_groups VALUES ('20000000-0000-0000-0000-000000000001');
INSERT INTO health_context_nodes VALUES ('30000000-0000-0000-0000-000000000001');
INSERT INTO notification_preferences(user_id,phone_number,sms_enabled) VALUES
 ('10000000-0000-0000-0000-000000000001','+256700000001',true),
 ('10000000-0000-0000-0000-000000000002','+256700000002',false),
 ('10000000-0000-0000-0000-000000000003',' ',true);
-- Inject an insert failure AFTER replacement has deleted and inserted rows.
-- Clients/users have no production FK; this trigger tests transaction atomicity.
CREATE FUNCTION fail_fixture_audience() RETURNS trigger AS $$
BEGIN
 IF to_jsonb(NEW)->>TG_ARGV[0] = 'ffffffff-ffff-ffff-ffff-ffffffffffff' THEN
  RAISE EXCEPTION 'injected audience insert failure';
 END IF;
 RETURN NEW;
END; $$ LANGUAGE plpgsql;
CREATE TRIGGER fail_client BEFORE INSERT ON announcement_clients
 FOR EACH ROW EXECUTE FUNCTION fail_fixture_audience('client_id');
CREATE TRIGGER fail_user BEFORE INSERT ON announcement_users
 FOR EACH ROW EXECUTE FUNCTION fail_fixture_audience('user_id');
