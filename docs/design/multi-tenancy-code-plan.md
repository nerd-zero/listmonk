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

**Status: implemented** (`internal/migrations/v6.4.0.go`, `schema.sql`,
issue #28). The original draft below this line in earlier versions of this
doc had two bugs, caught while grounding it against the actual `schema.sql`
and `queries/*.sql` before running anything — recorded here so a future
session doesn't reintroduce them:

1. **It dropped constraints `ON CONFLICT` clauses depend on.** The draft
   re-scoped `subscribers_email_key`, `links_url_key`, `users_username_key`/
   `users_email_key`, `templates_is_default_idx`, `campaigns_archive_slug_key`,
   and `roles`' `idx_roles`/`idx_roles_name` to composite `(tenant_id, ...)`
   forms in the *same* migration as the `tenant_id` column addition. But
   `queries/subscribers.sql` has `ON CONFLICT (email)`, `queries/links.sql`
   has `ON CONFLICT (url)`, and `queries/roles.sql` has
   `ON CONFLICT (parent_id, list_id)` — dropping the constraint these
   target breaks those upserts immediately, before the query-layer changes
   land (those are phase 4's job, per the "migrate users/roles uniqueness
   constraints" step in `multi-tenancy.md`'s phased plan — the high-level
   plan already had this right; only this file's detailed draft jumped
   ahead).
2. **The nullable → backfill → `SET NOT NULL` dance was unnecessary.**
   Adding the column as `NOT NULL DEFAULT 1 REFERENCES tenants(id)` in one
   statement covers existing rows *and* any row the (not-yet-tenant-aware)
   app inserts, since no `INSERT` in `queries/*.sql` sets `tenant_id` until
   a later phase threads it through. Postgres 11+ treats a constant
   `DEFAULT` on `ADD COLUMN` as metadata-only (no table rewrite), so this
   is simpler *and* removes a whole class of migration-ordering bugs.

**What actually shipped** is purely additive: a new `tenants` table, plus
one `tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE
CASCADE ON UPDATE CASCADE` column + one `idx_<table>_tenant` index on every
scoped table (the 9 originally listed) and every join/log table
(`subscriber_lists`, `campaign_lists`, `campaign_views`, `campaign_media`,
`link_clicks`) and `settings` — 15 tables total, all via `DEFAULT 1`, no
parent-join backfill logic needed since every row today belongs to tenant
1 regardless of table. **No existing constraint, index, or query was
touched.** See `internal/migrations/v6.4.0.go` for the exact SQL and
`schema.sql` for the fresh-install equivalent (the `tenants` table + seed
row is added first, since every other table's FK needs it to exist).

Explicitly deferred to the phase that also updates the corresponding
`queries/*.sql` file: re-scoping `subscribers_email_key`, `links_url_key`,
`users_username_key`/`users_email_key`, `templates_is_default_idx`,
`campaigns_archive_slug_key`, `idx_roles`, `idx_roles_name`, and
`settings_key_key`/the settings PK to composite `(tenant_id, ...)` forms —
that's phase 4 (auth/query threading) and phase 5 (settings), not phase 1.

Verified against the live dev DB: migration runs clean, is a no-op on
re-run (version-tracked), all 15 tables backfilled to tenant 1, and the
running app (Campaigns/Subscribers pages) was unaffected before and after.

---

## Phase 2 — RLS policies and indexing

**Status: implemented** (`internal/migrations/v6.5.0.go`, issue #29).

The original draft here (`USING (tenant_id = current_setting('app.current_tenant', true)::INTEGER)`,
"unset context sees nothing") had the same class of ordering bug as the
original Phase 1 draft: since the app doesn't set `app.current_tenant`
until the auth/request-flow phase, a strict policy would make **every**
query return zero rows the moment this migration runs against any
correctly-permissioned deployment (non-superuser, non-owner app role) —
a total outage, well before this dev sandbox (where the connecting role
happens to be a superuser) would ever show a symptom.

**What shipped instead** is deliberately permissive while tenant context is
unset:

```sql
ALTER TABLE subscribers ENABLE ROW LEVEL SECURITY;
ALTER TABLE subscribers FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON subscribers
    USING (
        tenant_id = current_setting('app.current_tenant', true)::INTEGER
        OR current_setting('app.current_tenant', true) IS NULL
    );
-- repeat per table (all 15 from phase 1: the 9 scoped tables, 5 join/log
-- tables, and settings)
```

Two deliberate additions beyond the original draft:

- **The `OR ... IS NULL` fallback** keeps the app fully functional today
  (single tenant, no session plumbing yet) while still exercising the real
  RLS machinery — the policy correctly restricts once a session sets
  `app.current_tenant` (verified below), it just doesn't punish the app for
  not setting it yet. **Tighten this** (drop the `OR` branch) once Phase 4
  lands and the app reliably sets tenant context on every request — tracked
  as a follow-up, not a new phase.
- **`FORCE ROW LEVEL SECURITY`** in addition to `ENABLE`: table owners are
  exempt from RLS by default, and most self-hosted listmonk installs
  (including this dev database) use a single Postgres role for both schema
  ownership and the app connection — so plain `ENABLE` alone would have
  silently been a no-op for that common deployment shape. `FORCE` closes
  that gap. Superusers remain exempt regardless of `FORCE` (no way around
  that in Postgres), which is why this is still inert in the current dev
  sandbox specifically — but it matters for any non-superuser owner role.

**Verify before merging this phase:**
```sql
-- confirm the app's role has no bypass
SELECT rolname, rolsuper, rolbypassrls FROM pg_roles WHERE rolname = current_user;
-- must show rolsuper=f, rolbypassrls=f
```
This dev database's role fails that check (superuser), which is why direct
verification required a throwaway non-owner, non-superuser role
(`rls_test_role`, `NOSUPERUSER NOBYPASSRLS`, granted table access, no RLS
bypass) — created, used to confirm all three cases (no context set → sees
all rows; `app.current_tenant='1'` → sees only tenant 1; `='2'` → sees only
tenant 2), then fully dropped afterward (role, grants, and the temporary
tenant-2 test data). Nothing from this test role persists in the repo or
the dev DB.

**Indexing:** the "tenant_id must be the leading column in composite
indexes" requirement from the RLS research notes above is satisfied by
Phase 1's per-table `idx_<table>_tenant` index (tenant_id alone, trivially
leading) — sufficient for the RLS predicate itself to use an index.
Rewriting *other* existing indexes (e.g. `idx_subs_status`,
`idx_camps_created_at`) to prepend `tenant_id` is a separate performance
optimization, not required for correctness, and is deferred until query-plan
profiling in a later phase shows an actual need — there's only one tenant
with data today, so there's nothing to profile against yet.

---

## Phase 3 — connection/session plumbing

**Status: implemented** (`internal/core/tenant.go`, `internal/core/tenant_test.go`,
issue #30). Two changes from the original draft:

1. **`WithTenant` is a `*Core` method, not a package-level function taking
   `*sqlx.DB`.** Phase 4 will call it from inside existing `Core` methods
   where `c` is already the receiver, so `c.WithTenant(ctx, tenantID, fn)`
   is more idiomatic than threading `c.db` through a free function — matches
   the existing style of other `Core` methods (`c.RefreshMatView`, etc.) in
   `internal/core/core.go`.
2. **The concurrency spike is a real, permanent test**
   (`TestWithTenant_ConcurrentIsolation`), not just a described plan — this
   was the repo's *first* Go test (`go test ./...` previously had zero test
   files anywhere outside `frontend/`), so it also had to establish how a
   Postgres-backed Go test connects to a database at all. It reads the same
   `LISTMONK_db__*` env vars `.github/workflows/tests.yml` already sets for
   CI, falling back to this repo's local dev `config.toml` values so it also
   runs against `make run`'s dev database with no extra setup, and skips
   (rather than fails) if no database is reachable.

Critically, the test does **not** run as the app's configured DB role.
Both this dev database's role and CI's default `postgres:16-alpine` image
role are superusers, and superusers (like table owners, see phase 2) bypass
RLS entirely regardless of policy — testing against either would make the
test pass for the wrong reason (or, if isolation were actually broken,
mask it). The test creates a throwaway, uniquely-named, least-privileged
role (`NOSUPERUSER NOBYPASSRLS`, granted only `SELECT`) for its own
duration, runs concurrent goroutines through `WithTenant` against **that**
role's connection pool, and tears the role + its temporary tenant/subscriber
test rows down in `t.Cleanup` afterward — mirroring the manual verification
process from phase 2.

```go
func (c *Core) WithTenant(ctx context.Context, tenantID int, fn func(tx *sqlx.Tx) error) error {
    tx, err := c.db.BeginTxx(ctx, &sql.TxOptions{})
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

Verified: `go test ./internal/core/... -run TestWithTenant_ConcurrentIsolation -race -count=1` passes reliably (run repeatedly to confirm — an early version had a `t.Cleanup`-ordering bug where an in-function `defer db.Close()` ran *before* the cleanup callback that still needed that connection, silently leaving orphaned test tenants/roles behind despite the test itself reporting `PASS`; fixed by closing the admin connection at the end of the same cleanup callback instead of via a separate `defer`). `go vet ./...` and the full `go test ./... -race` suite pass with the dev database up.

**Not in scope for this phase** (that's phase 4): no existing `Core` method
calls `WithTenant` yet. This is additive — a new file plus a new test, zero
behavior change to the running app, exactly like phases 1 and 2.

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
