-- Control-plane schema. The actual listmonk workspace each instance
-- provisions lives in the shared, multi-tenant listmonk fork's own
-- Postgres -- reached only through its Operator API, never this database.

CREATE TABLE users (
    id              uuid PRIMARY KEY,
    zitadel_subject text NOT NULL UNIQUE, -- OIDC "sub" claim from Zitadel
    email           text NOT NULL,
    display_name    text,
    -- Platform-wide access across every org's instances -- distinct from
    -- an org_members "owner" (scoped to one org) and from listmonk's own
    -- per-tenant "Super Admin" role. No self-service grant exists: set
    -- directly in the database by whoever operates the platform, the same
    -- deliberate-manual-action model as the listmonk fork's own static
    -- Operator API token.
    is_super_admin  boolean NOT NULL DEFAULT false,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

-- An org is the account-level entity that owns instances; a user reaches an
-- instance only via org membership, never directly. Every user gets a
-- personal org auto-created on first login (see internal JIT provisioning),
-- and can belong to more than one org via org_members.
CREATE TABLE orgs (
    id         uuid PRIMARY KEY,
    name       text NOT NULL,
    -- Mirrors this org 1:1 onto the listmonk fork's own cross-tenant
    -- Organization entity (an integer id there, not a uuid) -- created
    -- alongside this row so every instance under it can be grouped
    -- correctly on the listmonk side via organization_id.
    listmonk_organization_id integer UNIQUE,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE org_members (
    org_id     uuid NOT NULL REFERENCES orgs (id) ON DELETE CASCADE,
    user_id    uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role       text NOT NULL DEFAULT 'owner', -- 'owner' | 'member'
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (org_id, user_id)
);

CREATE INDEX org_members_user_id_idx ON org_members (user_id);

CREATE TABLE instances (
    id                 uuid PRIMARY KEY,
    org_id             uuid NOT NULL REFERENCES orgs (id) ON DELETE CASCADE,
    -- Same rule the Operator API enforces (cmd/operator.go's reTenantSlug in
    -- the listmonk fork). The workspace lives at {slug}.{listmonk root domain}.
    slug               text NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$'),
    name               text NOT NULL, -- workspace display name, passed straight through to the Operator API
    admin_username     text NOT NULL, -- passed to POST /api/operator/tenants and reused to reissue a setup link later
    admin_email        text NOT NULL,
    status             text NOT NULL DEFAULT 'created',
    -- listmonk's own tenants.id -- an integer there, not a uuid. Set once
    -- provision_listmonk_tenant succeeds.
    listmonk_tenant_id integer UNIQUE,
    -- Most recent one-time setup link from the Operator API. The token
    -- behind it lives in the listmonk fork's memory only and goes stale on
    -- its restart, so this is never treated as durable -- always re-issuable
    -- via POST /api/operator/tenants/{id}/setup-link ("resend setup link").
    admin_setup_url    text,
    created_at         timestamptz NOT NULL DEFAULT now(),
    updated_at         timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX instances_org_id_idx ON instances (org_id);

CREATE TABLE postmark_servers (
    id                  uuid PRIMARY KEY,
    instance_id         uuid NOT NULL UNIQUE REFERENCES instances (id) ON DELETE CASCADE,
    postmark_server_id  text NOT NULL,
    api_token_encrypted text NOT NULL,
    sending_domain      text NOT NULL,
    created_at          timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE dns_records (
    id            uuid PRIMARY KEY,
    instance_id   uuid NOT NULL REFERENCES instances (id) ON DELETE CASCADE,
    record_type   text NOT NULL, -- 'dkim' | 'return_path'
    host          text NOT NULL,
    value         text NOT NULL,
    verified      boolean NOT NULL DEFAULT false,
    created_at    timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX dns_records_instance_id_idx ON dns_records (instance_id);

-- Domain-level provisioning timeline shown in the UI (GET /instances/{id}/events).
-- Separate from River's own internal job/queue tables, which River's own
-- migration ("river migrate-up") manages and this schema does not touch.
CREATE TABLE provisioning_jobs (
    id          uuid PRIMARY KEY,
    instance_id uuid NOT NULL REFERENCES instances (id) ON DELETE CASCADE,
    job_type    text NOT NULL,
    status      text NOT NULL DEFAULT 'pending',
    attempts    integer NOT NULL DEFAULT 0,
    last_error  text,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX provisioning_jobs_instance_id_idx ON provisioning_jobs (instance_id);
