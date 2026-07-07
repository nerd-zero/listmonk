# Multi-tenancy: code-level implementation plan

Companion to [`multi-tenancy.md`](./multi-tenancy.md) (architecture/research) and
[`multi-tenancy-erd.drawio`](./multi-tenancy-erd.drawio) (schema diagrams). That
doc has the *why*; this one has the *how* — concrete SQL, Go signatures, and
file-by-file diffs to implement the 9 phases tracked as GitHub issues
[#28](https://github.com/nerd-zero/listmonk/issues/28)–[#35](https://github.com/nerd-zero/listmonk/issues/35)
and [#38](https://github.com/nerd-zero/listmonk/issues/38)
under [#27](https://github.com/nerd-zero/listmonk/issues/27).

Status: **not started**. Nothing below has been applied to the codebase yet.

---

## Phase 1 — schema foundation

### New migration: `internal/migrations/v6.4.0.go`

Follows the existing idempotent migration pattern (see `internal/migrations/v5.0.0.go`).
Register in `cmd/upgrade.go`'s `migList`:

```go
var migList = []migFunc{
    // ... existing entries ...
    {"v6.3.0", migrations.V6_3_0},
    {"v6.4.0", migrations.V6_4_0}, // multi-tenancy: schema foundation
}
```

```go
package migrations

import (
    "log"

    "github.com/jmoiron/sqlx"
    "github.com/knadh/koanf/v2"
    "github.com/knadh/stuffbin"
)

// V6_4_0 adds multi-tenancy scaffolding: a tenants table, a nullable
// tenant_id on every org-scoped table, and a backfilled default tenant so
// existing single-tenant installs keep working unmodified.
func V6_4_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
    if _, err := db.Exec(`
        DROP TYPE IF EXISTS tenant_status CASCADE;
        CREATE TYPE tenant_status AS ENUM ('active', 'suspended', 'disabled');

        CREATE TABLE IF NOT EXISTS tenants (
            id          SERIAL PRIMARY KEY,
            slug        TEXT NOT NULL UNIQUE,
            name        TEXT NOT NULL,
            status      tenant_status NOT NULL DEFAULT 'active',
            created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            updated_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        );

        -- Backfill: every existing install becomes tenant id=1.
        INSERT INTO tenants (id, slug, name, status)
        VALUES (1, 'default', 'Default', 'active')
        ON CONFLICT (id) DO NOTHING;
        SELECT setval('tenants_id_seq', GREATEST((SELECT MAX(id) FROM tenants), 1));
    `); err != nil {
        return err
    }

    // tenant_id added nullable + backfilled to 1, then set NOT NULL once
    // backfilled — this two-step avoids locking on ALTER ... NOT NULL
    // against a populated table in one shot.
    scopedTables := []string{
        "subscribers", "lists", "templates", "campaigns", "media",
        "links", "bounces", "roles", "users",
    }
    for _, t := range scopedTables {
        if _, err := db.Exec(`ALTER TABLE ` + t + ` ADD COLUMN IF NOT EXISTS tenant_id INTEGER REFERENCES tenants(id) ON DELETE CASCADE;`); err != nil {
            return err
        }
        if _, err := db.Exec(`UPDATE ` + t + ` SET tenant_id = 1 WHERE tenant_id IS NULL;`); err != nil {
            return err
        }
        if _, err := db.Exec(`ALTER TABLE ` + t + ` ALTER COLUMN tenant_id SET NOT NULL;`); err != nil {
            return err
        }
    }

    // Join/log tables: denormalize tenant_id directly rather than requiring
    // a join through the parent for RLS performance (leading-column index
    // requirement — see multi-tenancy.md's RLS research notes).
    joinTables := map[string]string{
        "subscriber_lists": "subscriber_id", // backfill via subscribers
        "campaign_lists":   "campaign_id",   // backfill via campaigns
        "campaign_views":   "campaign_id",
        "campaign_media":   "campaign_id",
        "link_clicks":      "campaign_id",
    }
    parentTable := map[string]string{
        "subscriber_id": "subscribers",
        "campaign_id":   "campaigns",
    }
    for t, fkCol := range joinTables {
        parent := parentTable[fkCol]
        if _, err := db.Exec(`ALTER TABLE ` + t + ` ADD COLUMN IF NOT EXISTS tenant_id INTEGER REFERENCES tenants(id) ON DELETE CASCADE;`); err != nil {
            return err
        }
        if _, err := db.Exec(`
            UPDATE ` + t + ` SET tenant_id = p.tenant_id
            FROM ` + parent + ` p WHERE p.id = ` + t + `.` + fkCol + ` AND ` + t + `.tenant_id IS NULL;
        `); err != nil {
            return err
        }
        // link_clicks/campaign_views/campaign_media allow NULL campaign_id
        // (row survives campaign deletion) — those rows keep tenant_id
        // NULL rather than forcing NOT NULL; RLS policy treats NULL as
        // "no tenant can see this" (safe default), see phase 2.
    }

    // settings: composite (tenant_id, key) — see phase 5 for the full
    // settings-split migration; this just adds the column here so schema
    // foundation is complete in one migration.
    if _, err := db.Exec(`
        ALTER TABLE settings ADD COLUMN IF NOT EXISTS tenant_id INTEGER REFERENCES tenants(id) ON DELETE CASCADE;
        UPDATE settings SET tenant_id = 1 WHERE tenant_id IS NULL;
        ALTER TABLE settings DROP CONSTRAINT IF EXISTS settings_key_key;
        ALTER TABLE settings ADD PRIMARY KEY (tenant_id, key);
    `); err != nil {
        return err
    }

    // Re-scope uniqueness constraints that were global.
    if _, err := db.Exec(`
        DROP INDEX IF EXISTS idx_subs_email;
        CREATE UNIQUE INDEX idx_subs_email ON subscribers(tenant_id, LOWER(email));

        ALTER TABLE campaigns DROP CONSTRAINT IF EXISTS campaigns_archive_slug_key;
        CREATE UNIQUE INDEX idx_camps_archive_slug ON campaigns(tenant_id, archive_slug);

        DROP INDEX IF EXISTS templates_is_default_idx;
        CREATE UNIQUE INDEX idx_templates_default ON templates(tenant_id, is_default) WHERE is_default = true;

        ALTER TABLE users DROP CONSTRAINT IF EXISTS users_username_key;
        ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;
        CREATE UNIQUE INDEX idx_users_username ON users(tenant_id, username);
        CREATE UNIQUE INDEX idx_users_email ON users(tenant_id, email);
    `); err != nil {
        return err
    }

    // Leading tenant_id composite indexes (RLS performance requirement).
    for _, t := range append(scopedTables, "subscriber_lists", "campaign_lists", "campaign_views", "campaign_media", "link_clicks") {
        if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_` + t + `_tenant ON ` + t + `(tenant_id);`); err != nil {
            return err
        }
    }

    return nil
}
```

### `schema.sql` updates

For fresh installs, every `CREATE TABLE` block for the tables above needs
`tenant_id INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE` added,
and the `tenants` table creation added before them (mirroring the migration).
The composite-unique-index changes above also need to replace the
single-column versions in `schema.sql` directly (don't keep both).

---

## Phase 2 — RLS policies and indexing

For every table in `scopedTables` + the five join tables + `settings`, a new
migration (`v6.5.0`) adds:

```sql
ALTER TABLE subscribers ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON subscribers
    USING (tenant_id = current_setting('app.current_tenant', true)::INTEGER);
-- repeat per table
```

The `true` second argument to `current_setting` makes it return `NULL`
instead of erroring when unset — important because migrations, the
`goyesql`-prepared statement setup in `cmd/init.go`, and any maintenance
script run outside a request context won't have `app.current_tenant` set.
Decide explicitly whether an unset tenant context means "see nothing" (NULL
never equals a tenant_id, so this is the default with the policy above — the
safe choice) vs. a superuser maintenance bypass path (a separate `BYPASSRLS`
maintenance role used only by migration/backup jobs, never by the app pool).

**Verify before merging this phase:**
```sql
-- confirm the app's role has no bypass
SELECT rolname, rolsuper, rolbypassrls FROM pg_roles WHERE rolname = current_user;
-- must show rolsuper=f, rolbypassrls=f
```

---

## Phase 3 — connection/session plumbing

New file: `internal/core/tenant.go`

```go
package core

import (
    "context"
    "database/sql"

    "github.com/jmoiron/sqlx"
)

// WithTenant runs fn inside a transaction with app.current_tenant set via
// SET LOCAL, so the setting is automatically cleared when the transaction
// ends — safe under a shared/pooled connection since SET LOCAL is
// transaction-scoped, not session-scoped.
func WithTenant(ctx context.Context, db *sqlx.DB, tenantID int, fn func(tx *sqlx.Tx) error) error {
    tx, err := db.BeginTxx(ctx, &sql.TxOptions{})
    if err != nil {
        return err
    }
    defer tx.Rollback()

    if _, err := tx.ExecContext(ctx, `SELECT set_config('app.current_tenant', $1::TEXT, true)`, tenantID); err != nil {
        return err
    }
    if err := fn(tx); err != nil {
        return err
    }
    return tx.Commit()
}
```

Before writing any handler code against this: spike a standalone test that
opens N goroutines against a shared `*sqlx.DB` pool, each calling
`WithTenant` with a different tenant ID and asserting the row set returned
never contains another tenant's rows, run under `-race`. `lib/pq`/`pgx`
connection-pool behavior with `SET LOCAL` inside explicit transactions
(rather than session-level `SET`) is the safe pattern per the RLS research
in `multi-tenancy.md` — this spike exists to prove it holds for this
specific driver/pool combination, not to re-derive the general advice.

---

## Phase 4 — auth and request-flow tenant resolution

### New file: `internal/tenant/resolve.go`

Subdomain-based resolution, inserted in `cmd/init.go:initHTTPServer`'s global
`srv.Use` (currently just `c.Set("app", app)`), running **before** `Auth.Middleware`:

```go
package tenant

import (
    "net"
    "net/http"
    "strings"
    "time"

    "github.com/labstack/echo/v4"
)

const CtxKey = "tenant"

// cache is a short-TTL slug->tenant lookup to avoid a DB round-trip per request.
var cache = newTTLCache(30 * time.Second)

func Middleware(core *core.Core, rootDomain string) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            host, _, _ := net.SplitHostPort(c.Request().Host)
            if host == "" {
                host = c.Request().Host
            }
            slug := strings.TrimSuffix(host, "."+rootDomain)
            if slug == host { // no subdomain present
                return echo.NewHTTPError(http.StatusNotFound)
            }

            t, ok := cache.Get(slug)
            if !ok {
                var err error
                t, err = core.GetTenantBySlug(slug)
                if err != nil {
                    return echo.NewHTTPError(http.StatusNotFound)
                }
                cache.Set(slug, t)
            }
            if t.Status != "active" {
                return c.Render(http.StatusOK, "tenant-unavailable.html", nil)
            }

            c.Set(CtxKey, t)
            return next(c)
        }
    }
}
```

### `internal/auth/models.go`

```go
type User struct {
    Base
    TenantID int `db:"tenant_id" json:"tenant_id"`
    // ... existing fields unchanged ...
}
```

`queries/users.sql`'s user-fetch queries need `tenant_id` added to their
`SELECT` lists (mirrors how `user_role_permissions` etc. are already
selected as joined columns).

### `internal/auth/auth.go`

`Auth.Middleware` (currently ~line 286-330) resolves a user from session/API
token and does `c.Set(UserHTTPCtxKey, user)`. Since the tenant middleware
above has already run and set `c.Get(tenant.CtxKey)`, add a cross-check
immediately after the existing `c.Set(UserHTTPCtxKey, user)` calls (lines 317
and 329):

```go
t := c.Get(tenant.CtxKey).(*models.Tenant)
if user.TenantID != t.ID {
    return echo.NewHTTPError(http.StatusForbidden)
}
c.Set(TenantIDHTTPCtxKey, user.TenantID)
```

This is defense-in-depth against a session/token issued on one tenant being
replayed against a different tenant's subdomain — see the host-only cookie
note below for the first layer.

**Session cookie:** set with no explicit `Domain` attribute (host-only,
scoped to `<slug>.listmonk.example.com` exactly) rather than a
`.listmonk.example.com`-wide cookie — a stolen cookie then can't even be
replayed against a different subdomain. Check wherever the session cookie is
currently written (`internal/auth` session store config) for an explicit
`Domain` and remove it if present.

For **public** unauthenticated routes (unsubscribe, archive, tracking pixel —
see phase 8), the tenant middleware above has *already* resolved tenant from
the subdomain — handlers filter their entity lookups by that resolved
`tenant_id` rather than deriving tenant from the entity's own row.

### OIDC callback (knock-on effect of per-tenant settings, phase 5)

The OIDC callback route must run after tenant-resolution middleware (already
true, since it's global) so it can load `GetSettings(tenantID).OIDC` for
*this* tenant before validating the auth-code exchange — each tenant has its
own IdP client ID/secret/issuer.

### `internal/core/*` — threading `tenantID`

Every entry point gets a leading `tenantID int` parameter, mirroring the
existing `getAll bool, permittedIDs []int` list-permission pattern. Example
diff for `internal/core/subscribers.go`:

```go
// before
func (c *Core) QuerySubscribers(searchStr, queryExp string, listIDs []int, subStatus string, order, orderBy string, offset, limit int) (models.Subscribers, int, error) {

// after
func (c *Core) QuerySubscribers(tenantID int, searchStr, queryExp string, listIDs []int, subStatus string, order, orderBy string, offset, limit int) (models.Subscribers, int, error) {
```

Internally, wrap the existing query execution in `core.WithTenant(ctx, c.db, tenantID, func(tx *sqlx.Tx) error { ... })`
from phase 3, so RLS enforces the boundary — the `tenantID` parameter here is
belt-and-suspenders (makes intent explicit in the Go call site, catches bugs
in tests without a real Postgres RLS setup), not the sole enforcement layer.

This same signature change applies to every exported `Core` method in:
`subscribers.go`, `lists.go`, `campaigns.go`, `templates.go`, `media.go`,
`bounces.go`, `subscriptions.go`, `users.go`, `roles.go`, `settings.go`,
`dashboard.go` — all callers in `cmd/*.go` handlers need the extra argument
threaded from `auth.GetUser(c).TenantID`.

### `internal/auth/models.go` — super-admin scope

`SuperAdminRoleID = 1` currently short-circuits every permission check
cross-request. Needs to become tenant-scoped: a super admin is admin *within
their tenant*, not across tenants. If a cross-tenant instance-operator role
is wanted (open question in `multi-tenancy.md`), it should be a distinct new
concept (e.g. `user.Type == "operator"`) checked separately, not reuse
`SuperAdminRoleID`.

---

## Phase 5 — settings, fully per-tenant

**Decided:** no global/per-tenant split. Every key — including `smtp`,
`security.oidc`, and `upload.s3.*` — is per-tenant. `models.Settings` stays a
single flat struct; only the load/save path gains a `tenantID` parameter.

### `internal/core/settings.go`

```go
// before
func (c *Core) GetSettings() (models.Settings, error) {
func (c *Core) UpdateSettings(s models.Settings) error {

// after
func (c *Core) GetSettings(tenantID int) (models.Settings, error) {
func (c *Core) UpdateSettings(tenantID int, s models.Settings) error {
```

`queries/settings.sql`'s `get-settings`/`update-settings` need a `tenant_id`
parameter added to their `WHERE`/`INSERT ... ON CONFLICT` clauses — one query
pair, not two, since there's no global/tenant split to maintain.

### Knock-on effects to implement alongside this phase

- **`internal/media`'s S3 client**: currently constructed once at startup
  from global settings. Change to a per-tenant client cache (map keyed by
  `tenant_id`, lazily constructed on first use, same TTL-cache shape as the
  tenant-slug cache in phase 4) since each tenant now supplies its own
  bucket/credentials.
- **OIDC callback**: see phase 4 — needs tenant resolved (from subdomain)
  before it can load the right IdP config to validate against.

---

## Phase 6 — tenant-aware manager/dispatcher

### `internal/manager/manager.go`

`Store` interface (`~line 50`) — `NextCampaigns` needs tenant awareness:

```go
// before
NextCampaigns(excludeIDs []int64, largestID int) (models.Campaigns, error)

// after — scan tenant-by-tenant rather than globally
NextCampaignsForTenant(tenantID int, excludeIDs []int64, largestID int) (models.Campaigns, error)
```

`scanCampaigns` (`~line 447`) changes from a single global scan to iterating
active tenant IDs (a new cheap `SELECT id FROM tenants WHERE status='active'`
call each tick, cached) and calling `NextCampaignsForTenant` per tenant,
setting `app.current_tenant` for that batch via `core.WithTenant`. SMTP pool
construction (currently built once globally from the global `smtp` setting)
needs to become per-tenant too, sourced from `GetTenantSettings(tenantID).SMTP`
— this is the part of this phase most coupled to phase 5's settings split
landing first.

Rate limiting (`app.message_rate`/`app.concurrency`) similarly needs a
decision: keep one global worker-pool size shared across all tenants'
campaigns (simplest, no per-tenant starvation protection) vs. per-tenant
worker pools (fairer, more moving parts). Recommend starting with global
pool + per-tenant send-rate throttling only, revisit if a tenant experiences
starvation in practice.

---

## Phase 7 — frontend

No tenant switcher needed under the fully-isolated-orgs model (one user, one
tenant). Changes are audit-only:

- `frontend/src/store/index.js`: confirm no cached list/campaign/subscriber
  IDs persist across a login-as-different-tenant-user flow (e.g. logout
  should clear the Pinia store, not just redirect).
- `frontend/src/api/generated/`: no changes needed — tenant scoping happens
  server-side; the client never sends or needs a tenant ID explicitly.

---

## Phase 8 — public-facing route audit

Routes to check in `cmd/handlers.go`'s public group: unsubscribe page,
subscription-preferences page, campaign archive view, opt-in confirmation,
link-click redirect, tracking pixel. Pattern for each: the handler currently
does something like

```go
sub, err := app.core.GetSubscriber(0, subUUID, "")
```

Post-migration, `GetSubscriber` (and equivalent lookups) must resolve
tenant *from* the UUID's own row rather than requiring it as an input — i.e.
`SELECT tenant_id FROM subscribers WHERE uuid = $1` is itself the tenant
resolution step for public routes, then subsequent joins (e.g. rendering
list name, campaign sender branding) must filter by that resolved
`tenant_id`, not trust a second UUID/ID param without checking it belongs to
the same tenant as the first. This is the concrete thing an audit test should
assert: pass a campaign UUID from tenant A alongside a list ID from tenant B
and confirm the response 404s rather than mixing data.

---

## Phase 9 — operator API

New route group, entirely separate from the tenant-facing app: no session/JWT
auth, no RLS-scoped DB role.

### `internal/operator/` (new package, mirrors `internal/auth/` structure)

```go
package operator

// Middleware checks a static bearer token from config against every request
// to /api/operator/*. No per-operator identity in v1 — see multi-tenancy.md's
// decisions log for why (audit trail is a known gap, acceptable for v1).
func Middleware(token string) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            got := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")
            if subtle.ConstantTimeCompare([]byte(got), []byte(token)) != 1 {
                return echo.NewHTTPError(http.StatusUnauthorized)
            }
            return next(c)
        }
    }
}
```

Config: `config.toml` gains an `[operator]` section (`token = "..."`), or
`LISTMONK_operator__token` env var, following the existing `LISTMONK_`
double-underscore convention.

### DB role

A new Postgres role (e.g. `listmonk_operator`) with `BYPASSRLS`, distinct
from the tenant-app pool's role (which must never have it — see phase 2's
verification query). This is the second consumer of a bypass role, alongside
migrations/backup jobs.

### Routes (`cmd/handlers.go`, new group under `/api/operator`)

```go
g := srv.Group("/api/operator", operator.Middleware(app.constants.OperatorToken))
g.GET("/tenants", handleOperatorListTenants)
g.GET("/tenants/:id", handleOperatorGetTenant)
g.POST("/tenants", handleOperatorCreateTenant)
g.PUT("/tenants/:id/status", handleOperatorUpdateTenantStatus)
```

- `handleOperatorListTenants` / `GetTenant`: tenant row + basic counts
  (`SELECT COUNT(*) FROM users WHERE tenant_id = $1`, same for subscribers)
  for a support/billing dashboard.
- `handleOperatorCreateTenant`: inserts a `tenants` row + an initial admin
  `users` row in one transaction. **Open question (see multi-tenancy.md):**
  how the initial admin's password/invite is communicated — placeholder
  behavior for v1 is to return a one-time setup-link token in the response
  body, expiring after first use, rather than emailing anything (email
  requires that tenant's own SMTP config, which doesn't exist yet for a
  brand-new tenant).
- `handleOperatorUpdateTenantStatus`: validates the new status is one of
  `active`/`suspended`/`disabled`, updates `tenants.status`. Suspended/disabled
  tenants: the phase-4 resolution middleware renders the "workspace
  unavailable" page for any request on their subdomain, and phase 6's
  `scanCampaigns` skips them on the next tick (no immediate kill of an
  in-flight send — same "finish current batch, don't start new ones"
  semantics as pausing a single campaign today).

---

## Testing strategy across all phases

- Unit: every `internal/core` function gets a table-driven test with ≥2
  tenants seeded, asserting tenant B's data never appears in tenant A's
  query results.
- Integration: the phase 3 concurrency spike becomes a permanent test in
  `internal/core/tenant_test.go`, run in CI.
- Migration: `internal/migrations/v6.4.0_test.go` — apply migration to a
  seeded pre-migration single-tenant DB fixture, assert all rows land in
  tenant 1 and constraints are correctly re-scoped.
