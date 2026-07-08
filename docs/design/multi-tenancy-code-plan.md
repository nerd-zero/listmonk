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

**Status: subdomain resolution implemented (issue #31); `internal/core`
tenantID-threading split out to a new follow-up issue.** Investigation
while grounding this phase found the `internal/core` part alone touches
**107** exported `Core` methods and **150** call sites in `cmd/*.go`, none
of which take a `context.Context` today and none run through a
transaction — they call directly on pool-level prepared statements
(`c.q.X.Select(...)`). Making phase 3's `WithTenant` actually apply to
real queries means rebinding each one via `tx.Stmtx(c.q.X)`, a mechanical
but invasive change across all 107 methods (plus adding `ctx` to all of
them as a prerequisite, since none have it). Too large to do safely in
one pass alongside the resolution work — deliberately deferred, tracked as
a separate follow-up rather than attempted piecemeal.

**What shipped** is the resolution/auth half only, gated behind a new
`app.multi_tenancy_enabled` config flag (default `false`) — purely
additive, like phases 1-3.

### `models/tenant.go` (new)

Holds `type Tenant struct { Base; Slug, Name, Status string }` *and* the
shared context-key constant `TenantCtxKey`. Both live in `models`, not in
`internal/tenant`, because of an import-cycle constraint discovered here:
`internal/core` already imports `internal/auth` (for permission types used
in `subscribers.go`/`users.go`/`roles.go`). A tenant-resolution middleware
needs `*core.Core`, so it can't live in — or be imported by —
`internal/auth`, or you get `auth → tenant → core → auth`. Putting the
struct and context key in `models` (which has zero internal dependencies)
lets both `internal/tenant` (imports `core` + `models`) and
`internal/auth` (imports only `models`) use them without cycling.

### `internal/tenant/resolve.go` (new)

```go
func Middleware(core *core.Core, rootDomain string, enabled bool) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            if !enabled {
                c.Set(models.TenantCtxKey, defaultTenant) // stub, ID: 1, no DB hit
                return next(c)
            }

            host := c.Request().Host
            if h, _, err := net.SplitHostPort(host); err == nil {
                host = h
            }
            slug := strings.TrimSuffix(host, "."+rootDomain)
            if slug == host || slug == "" {
                return echo.NewHTTPError(http.StatusNotFound)
            }

            t, err := core.GetTenantBySlug(slug)
            if err != nil {
                return echo.NewHTTPError(http.StatusNotFound)
            }
            if t.Status != "active" {
                return echo.NewHTTPError(http.StatusServiceUnavailable, "workspace unavailable")
            }

            c.Set(models.TenantCtxKey, &t)
            return next(c)
        }
    }
}
```

No slug→tenant cache yet (the original draft's TTL cache) — deferred the
same way phase 2 deferred composite-index rework: there's only ever one
tenant to look up right now, revisit if profiling shows a need once
there's more than one.

Registered in `cmd/init.go:initHTTPServer`'s global `srv.Use`, immediately
after the existing `c.Set("app", app)` middleware and before
`initHTTPHandlers(srv, app)` sets up the app's 3 separate route groups
(authenticated pages, authenticated `/api`, public) — confirmed this is
the *only* middleware that spans all 3 today, which is exactly why it's
the right insertion point for something that must also cover public
routes (unsubscribe, archive, tracking pixel).

### `internal/auth/models.go`

```go
type User struct {
    Base
    TenantID int `db:"tenant_id" json:"tenant_id,omitempty"`
    // ... existing fields unchanged ...
}
```

No query changes needed: `queries/users.sql`'s `get-user`/`get-users` both
already `SELECT *`/`users.*`, so `tenant_id` (a real column since phase 1)
was already coming back in every result row.

### `internal/auth/auth.go`

`Auth.Middleware` never aborts on failure itself — on any failure it does
`c.Set(UserHTTPCtxKey, echo.NewHTTPError(...))` and calls `next(c)` anyway;
rejection is deferred to per-route-group wrapper middleware
(`redirectIfUnauth`/`jsonErrorIfUnauth` in `cmd/handlers.go`) that inspects
`c.Get(UserHTTPCtxKey)`. The tenant cross-check follows this exact
convention rather than returning an error directly:

```go
func tenantMismatch(c echo.Context, user User) bool {
    t, ok := c.Get(models.TenantCtxKey).(*models.Tenant)
    return ok && user.TenantID != t.ID
}
```
called after both places a user is successfully resolved (API-token path
and session path), setting the *same* `"invalid session"` error the
existing invalid-session case uses — so a cross-tenant replay is
indistinguishable from an expired session to the client, no extra
information disclosed.

Verified end-to-end against a temporary second backend instance
(`LISTMONK_app__multi_tenancy_enabled=true LISTMONK_app__root_domain=localhost`,
different port, same dev DB): logged in on tenant 1's own subdomain
(`default.localhost`), confirmed the resulting session cookie works there,
then confirmed reusing that exact cookie against a second throwaway
tenant's subdomain gets rejected with `403 invalid session` — the concrete
"session issued for tenant A replayed against tenant B" scenario this
check exists for.

**Session cookie:** already host-only (no `Domain` attribute set anywhere
in `internal/auth`'s `simplesessions.Options` or `cmd/init.go`'s
`initAuth()`) — this requirement was already satisfied before this phase,
nothing to change.

### OIDC callback, `internal/core/*` tenantID-threading, super-admin scope

Deferred — see status note above. OIDC callback tenant-awareness depends
on phase 5 (per-tenant settings) landing first; `internal/core` threading
is its own follow-up issue given the 107-method scope; `SuperAdminRoleID`
becoming tenant-scoped depends on that threading landing.

---

## Phase 5 — settings, fully per-tenant

**Decided:** no global/per-tenant split. Every key — including `smtp`,
`security.oidc`, and `upload.s3.*` — is per-tenant. `models.Settings` stays a
single flat struct; only the load/save path gains a `tenantID` parameter.

**Status: DB layer implemented (issue #32); subsystem redesign split into
its own follow-up.** Investigation before starting found this phase bundles
two very different sizes of work, the same pattern as phase 4/issue #40:
settings aren't read per-request — they're loaded **once at process boot**
into a global `koanf` config, and the SMTP messenger pools, the media/S3
store, the OIDC config, and the campaign manager are all built **once as
process-lifetime singletons** from that global state (`cmd/main.go`'s
`initSMTPMessengers`/`initMediaStore`/`initAuth`/`initCampaignManager`).
There's no live-reload mechanism today at all — even the *existing*
single-tenant "update settings" flow requires a full process re-exec
(`cmd/settings.go`'s `handleSettingsRestart` → `syscall.Exec`) to take
effect. Making those four subsystems genuinely per-tenant is a redesign,
not a parameter addition — split into its own follow-up issue, same as
`internal/core` threading was split into #40. **What shipped this session
is the DB/`Core` layer only**, with those subsystems deliberately left as
global singletons pinned to tenant 1 (documented in code, zero behavior
change from today).

### Migration `v6.6.0`: composite `(tenant_id, key)` primary key

Re-scopes `settings` from the phase-1-deferred `UNIQUE(key)` to a real
composite key — this was explicitly left for "the phase that also updates
the corresponding `queries/*.sql` file," and this is that phase:

```sql
ALTER TABLE settings DROP CONSTRAINT IF EXISTS settings_key_key;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid = 'settings'::regclass AND contype = 'p') THEN
        ALTER TABLE settings ADD PRIMARY KEY (tenant_id, key);
    END IF;
END $$;
```

**Found and fixed while running this migration, not before:** `cmd/install.go`'s
`recordMigrationVersion` — the upgrade runner's *own* bookkeeping, used by
every migration including this one — does `INSERT INTO settings (key,
value) VALUES ('migrations', ...) ON CONFLICT (key) DO UPDATE ...`. Dropping
`settings_key_key` broke it immediately (confirmed by actually running the
migration, not just reading the code): `pq: there is no unique or exclusion
constraint matching the ON CONFLICT specification`. Fixed by adding
`tenant_id` to the insert and changing to `ON CONFLICT (tenant_id, key)`.
Also fixed `cmd/upgrade.go`'s `getLastMigrationVersion` (`SELECT value->>-1
FROM settings WHERE key='migrations'`) to add `AND tenant_id=1` for the same
reason — both pinned to tenant 1, matching `initSettings` below. This is
the exact class of bug phase 1 was originally trying to avoid by deferring
constraint changes; it still slipped through because this specific query
lives in the *migration framework itself*, not `queries/*.sql`, so it
wasn't caught by grepping for `ON CONFLICT` across the query files during
planning.

### `internal/core/settings.go`

```go
func (c *Core) GetSettings(ctx context.Context, tenantID int) (models.Settings, error)
func (c *Core) UpdateSettings(ctx context.Context, tenantID int, s models.Settings) error
func (c *Core) UpdateSettingsByKey(ctx context.Context, tenantID int, key string, value json.RawMessage) error
```
All three now run through `WithTenant`, using the `stmtx()` helper from
issue #40's slice 1 (same `.Unsafe()`-preservation requirement applies
here). `queries/misc.sql`'s `get-settings`/`update-settings`/
`update-settings-by-key` all gained a `tenant_id` parameter. For
`update-settings` specifically, the `tenant_id` filter is **required for
correctness, not just defense-in-depth**: it does `WHERE s.key = c.key` for
each key in an incoming JSON map, and since key names now repeat across
tenants by design, an unfiltered version would silently update every
tenant's row sharing a key name in that map.

### `cmd/init.go`'s `initSettings` — pinned to tenant 1

The boot-time global-config load now explicitly passes tenant 1 to
`get-settings`'s new `$1` parameter, with a code comment explaining this is
intentional and matches the "subsystems stay global singletons" scoping
decision above — not an oversight to fix later independently of the
subsystem-redesign follow-up.

### Knock-on effects — now explicitly deferred to the subsystem-redesign follow-up

- **`internal/media`'s S3 client**: per-tenant client cache (map keyed by
  `tenant_id`, lazily constructed).
- **OIDC callback**: needs tenant resolved before loading IdP config (phase 4
  already flagged this).
- **SMTP messenger pools**: per-tenant construction instead of one global
  pool built at boot.
- **Live settings reload**: replacing `syscall.Exec` full-process-restart
  with something that can refresh one tenant's config without affecting
  every other tenant's running state.

---

## Phase 6 — tenant-aware manager/dispatcher

**Status: implemented.** Was blocked on real per-tenant dispatch existing
to receive tenant-scoped scan results — unblocked once #41 slice 1 (SMTP
messenger resolution, below) shipped.

### What shipped

`Store` interface gained `NextCampaigns(tenantID int, currentIDs,
sentCounts []int64) ([]*models.Campaign, error)` (tenantID added as the
leading param) and a new `GetActiveTenantIDs() ([]int, error)` method,
backed by a plain `SELECT id FROM tenants WHERE status = 'active'` query
(no caching added — this only ever returns one row today with no way to
create more, so premature to optimize; revisit if profiling shows a need,
same deferral pattern used elsewhere in this plan).

`scanCampaigns` now: fetches active tenant IDs each tick, groups the
in-flight campaigns already being tracked (`m.pipes`) by tenant (via the
`TenantID` field `models.Campaign` gained in #41 slice 1), then calls
`NextCampaigns` once per active tenant with only that tenant's in-flight
IDs/counts. `queries/campaigns.sql`'s `next-campaigns` gained a `tenant_id
= $3` filter on its `camps` CTE.

**Why "group in-flight campaigns by tenant" rather than just adding the
tenant filter and calling it once per tenant with the same global
in-flight list each time:** `next-campaigns` reuses its `$1` (current IDs)
parameter for two purposes — excluding those campaigns from re-selection,
*and* (via `unnest($1::INT[], $2::INT[])` in an `updateCounts` CTE)
incrementing their `sent` counts in the DB. Passing the same global list to
every per-tenant call would apply that increment once per tenant scanned
each tick, double/triple/etc.-counting `sent`. Traced through this before
writing any code, not discovered by testing — the fix is in
`internal/manager`'s `getCurrentCampaigns` (now groups into
`map[int]currentCampaignIDs`), not in the SQL.

### A real bug caught by live-testing, not by reading the code

The zero-value of `currentCampaignIDs` (returned by a map lookup for a
tenant with zero in-flight campaigns — the common case) has `nil` `ids`/
`counts` slices. `pq.Int64Array(nil).Value()` serializes to SQL `NULL`,
not an empty array (confirmed via a throwaway Go program, not assumed) —
and `NOT(campaigns.id = ANY(NULL::INT[]))` evaluates to SQL `NULL` under
ordinary three-valued comparison semantics, which a `WHERE` clause treats
as "false", filtering out **every** row. A campaign set to `running` was
never picked up; no error anywhere, just silently zero results every
tick. Root-caused by comparing against the original single-tenant
`getCurrentCampaigns`, which explicitly built its slices with
`make([]int64, 0, len(m.pipes))` rather than a zero-value default — a
detail its own comment ("needs to return an empty slice in case there are
no campaigns") called out, that got lost when restructuring the return
type to a per-tenant map. Fixed by explicitly defaulting `e.ids`/
`e.counts` to `[]int64{}` before each `NextCampaigns` call, with a comment
explaining why so the next refactor of this code doesn't reintroduce it.

**Verified live**, not just built: started the existing draft campaign via
the API, confirmed via the backend log that `scanCampaigns` picked it up
on the very next tick, `getMessenger` resolved tenant 1's SMTP messenger
(from #41 slice 1), and the send failed at the expected point (a real
timeout against this dev environment's fake mail host) — then confirmed
the fix by reproducing the *bug* first (campaign stuck in `running`,
zero ticks logged as picking it up, across several tick cycles) before
applying and re-verifying the fix. Campaign state reset to `draft`
afterward.

Rate limiting (`app.message_rate`/`app.concurrency`) still shares one
global worker-pool size across all tenants' campaigns — deferred, matching
the original draft's recommendation to start simple and revisit only if a
tenant experiences starvation in practice.

---

## Issue #41 — per-tenant SMTP/media/OIDC/manager (4 subsystems, same
shape as #40)

### Slice 1 — SMTP messenger resolution: implemented

Investigation found `internal/manager` never imports `internal/core` (its
`Store` interface, implemented by `cmd/manager_store.go`, already bridges
DB access without one) and neither it nor `internal/messenger/email`
imports the other — confirmed via grep before designing, not assumed, so
wiring a new per-tenant resolver from `cmd/` carries zero cycle risk.

**What shipped:**
- `models.Campaign` and `models.Message` gained a `TenantID` field (the
  former free via existing `campaigns.*` queries, same pattern as
  `auth.User.TenantID`; the latter explicit, since the transactional send
  path — `cmd/tx.go`'s `SendTxMessage` — has no `Campaign` to derive it
  from).
- New `internal/manager.MessengerResolver` interface (`GetMessenger(ctx,
  tenantID, name) (Messenger, error)`, plus an `ErrMessengerNotFound`
  sentinel) and a `Manager.getMessenger` helper that tries the resolver
  first, falling back to the existing process-global `m.messengers` map
  (unchanged) on `ErrMessengerNotFound` — this is what keeps postback
  messengers (a different messenger type, out of scope for this slice)
  working without needing to explicitly classify messenger types.
- Concrete resolver: `cmd/tenant_messenger.go`'s `tenantMessengers`, holding
  a `*core.Core` reference and a `map[int]map[string]manager.Messenger`
  cache (mutex-protected, populated lazily, **never invalidated within a
  process's lifetime** — deliberately matching the existing
  process-global messenger set's staleness contract, since settings
  updates already require a full `syscall.Exec` restart to take effect
  today, confirmed live during phase 5. Live invalidation is out of scope,
  tracked as part of this issue's remaining work).
- `cmd/init.go`'s `initSMTPMessengers()` refactored to accept a
  `*koanf.Koanf` parameter (instead of closing over the global `ko`) and
  to **return an error instead of calling `lo.Fatalf`** — this was a
  necessary change beyond the original plan, found while implementing:
  the function is now called from two contexts (boot, which should still
  fail fast on bad config, vs. the lazy per-tenant path, where crashing
  the whole process over one tenant's malformed SMTP config would take
  down every tenant). The boot-time caller (`cmd/main.go`) still calls
  `lo.Fatalf` itself on the returned error, preserving today's behavior
  exactly.
- The per-tenant resolver reuses `initSMTPMessengers` rather than
  reimplementing SMTP-config parsing: it round-trips a tenant's
  `models.Settings` through `json.Marshal`/`json.Unmarshal` into a fresh
  per-tenant `*koanf.Koanf` (the JSON tags on `Settings` already match the
  dotted keys `initSettings` loads from the DB at boot), then calls the
  same function. This avoids a naive JSON-only conversion into
  `email.Server`, which embeds `smtppool.Opt` via a mapstructure-only
  `,squash` tag that plain `encoding/json` doesn't understand.

**Verified live, not just built/tested:** started the existing draft "Test
campaign" via the API and watched the backend log show
`initSMTPMessengers`'s init line fire a *second* time (the first is the
boot-time global construction; the second, at campaign-start, is the lazy
per-tenant resolver building tenant 1's messenger set for the first time),
followed by `newPipe`'s validation passing (no "unknown messenger" error)
and a real SMTP connection attempt that failed with a legitimate network
timeout against this dev environment's fake mail host — the expected,
correct failure mode, proving resolution succeeded and only delivery
(irrelevant to this slice) failed. Campaign state reset to `draft`
afterward to restore the dev DB baseline.

**Remaining in #41, not this slice:** media/S3 per-tenant client cache,
OIDC per-tenant config resolution, and Phase 6's scan-side tenant
awareness (now unblocked, see above).

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

## Issue #40 — threading tenantID through internal/core

Split out from phase 4 (see that phase's status note) once investigation
found it touches 107 exported `Core` methods / 150 call sites, none of
which take a `context.Context` or run through a transaction today. Being
done file-by-file; each slice's findings are recorded here so the next
slice starts informed rather than re-discovering the same issues.

### Slice 1 — `subscribers.go` (issue #40, first slice): implemented

Confirmed the 3 execution shapes anticipated when scoping this work (see
this file's earlier "Tenant ID source at call sites" note under phase 4):
simple prepared-statement calls (rebind via a `stmtx` helper, below),
self-managed raw-SQL transactions (replace the method's own `BeginTxx`
with `WithTenant`), and one long-lived iterator (`ExportSubscribers`, left
with a `TODO(#40)` per the deferred-scoping decision).

**Critical finding, applies to every future slice:** `sqlx.Tx.Stmtx()`
does **not** propagate the `.Unsafe()` flag. `cmd/init.go` opens the pool
`*sqlx.DB` with `.Unsafe()` (several models don't map every returned
column — `internal/core/campaigns.go` already has a comment about this for
`models.Campaigns`, and now *every* table has an unmapped `tenant_id`
column since phase 1 unless its model struct happens to declare one).
`BeginTxx` correctly copies `.Unsafe()` onto the resulting `*sqlx.Tx`
(confirmed in sqlx v1.4.0 source: `&Tx{..., unsafe: db.unsafe, ...}`), but
`Tx.Stmtx()` does not carry it onto the `*sqlx.Stmt` it derives
(`&Stmt{Stmt: tx.Stmt(s), Mapper: tx.Mapper}` — no `unsafe` field set).
Symptom: `missing destination name tenant_id in *models.X` errors that
only appear on the tx-rebound path — `tx.Select(...)`/`tx.Get(...)` called
*directly* (shape 2) work fine since they inherit `.Unsafe()` correctly;
only statements rebound via `Stmtx` (shape 1) hit this.

**Fix, already in place for all future slices to reuse:** `internal/core/tenant.go`
now has a package-level `stmtx(tx *sqlx.Tx, stmt *sqlx.Stmt) *sqlx.Stmt`
helper that wraps `tx.Stmtx(stmt).Unsafe()`. **Use `stmtx(tx, c.q.X)`, never
`tx.Stmtx(c.q.X)` directly**, for every future slice's shape-1 methods.

**Other changes needed for shape 2 and the shared query-template helpers:**
- `Core.WithTenant`'s signature gained a `*sql.TxOptions` parameter (`nil`
  = `BeginTxx`'s own default) so shape-2 methods that need
  `&sql.TxOptions{ReadOnly: true}` as a security control (subscribers.go's
  arbitrary-query features — see the code comment there) don't lose that
  guarantee to the helper.
- `models/queries.go`'s `compileSubscriberQueryTpl`/`ExecSubQueryTpl`
  (shared by `subscribers.go` and `subscriptions.go`) had their `db
  *sqlx.DB` parameter widened to `db sqlx.Execer` (satisfied by both
  `*sqlx.DB` and `*sqlx.Tx`), and `compileSubscriberQueryTpl`'s own nested
  `BeginTxx`/`Rollback` was removed (confirmed safe: its "dry run" flag
  just changes a `LIMIT` clause on a pure `SELECT`, no mutation possible
  either way) — a transaction can't `BeginTxx` a second time inside itself.
  Migrated callers (`subscribers.go`) now pass the open `tx`; unmigrated
  callers (`subscriptions.go`, not yet in scope) keep passing `c.db`
  unchanged and are unaffected.
- **Tenant ID at call sites**: settled on always reading the
  subdomain-resolved tenant (`c.Get(models.TenantCtxKey)`, via a new
  `tenantID(c echo.Context) int` helper in `cmd/handlers.go`) rather than
  `auth.GetUser(c).TenantID` — simpler (one source, not two depending on
  route type) and already proven consistent by phase 4's auth cross-check.

**Verified:** full CRUD (create/view/edit/blocklist/delete, both
individual and bulk), the advanced arbitrary-query search path, and CSV
export all work identically under the default (multi-tenancy disabled)
config — tested live via the browser and `curl`. Real cross-tenant
isolation was verified in two layers: (1) confirmed **RLS itself** blocks
cross-tenant rows when queried via a non-superuser role with
`app.current_tenant` set (same throwaway-role technique as phase 2/3);
(2) confirmed the full HTTP path (two tenants, two logged-in users, real
subdomains) does **not** show isolation in this specific dev environment,
but only because — as already flagged in phase 2 — the dev DB role
(`listmonk-dev`) is a superuser and superusers always bypass RLS
regardless of policy. This is a pre-existing, already-documented
dev-environment limitation, not a defect in this slice's code; the
per-layer test above is what actually confirms the code is correct.
**Any real deployment of this feature must use a non-superuser,
non-table-owner application role**, per the RLS "verify before merging"
checklist in phase 2.

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
