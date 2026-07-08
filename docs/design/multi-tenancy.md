# Multi-tenancy: research and implementation plan

Status: **phases 1-3 and 6 implemented; phase 4 partially implemented
(auth/subdomain resolution shipped; `internal/core` tenantID-threading,
split into its own follow-up issue #40, is now fully complete —
`subscribers.go`, `subscriptions.go`, `campaigns.go`, `bounces.go`,
`lists.go`, `media.go`, `roles.go`, `templates.go`, `users.go`, and
`dashboard.go` are all threaded through `WithTenant`/RLS, including two
cross-cutting fixes found while closing it out: every `INSERT` across
the schema now sets `tenant_id` explicitly (previously relied on
`DEFAULT 1`, which a real non-superuser RLS role would have rejected for
any tenant but 1), and the three dashboard/subscriber-count materialized
views now carry a `tenant_id` dimension (migration `v6.7.0`) after a live
cross-tenant read leak was found in the global fallback row they used to
share — see Decisions log for both); phase 5 fully implemented (settings
DB/Core layer shipped, subsystem redesign — SMTP/media/OIDC/manager —
split into its own follow-up issue #41, now fully complete: SMTP,
media/S3, and OIDC per-tenant resolution all shipped); phase 8 (public
route audit) done — no cross-tenant leak found in public routes, but the
audit surfaced and fixed 3 severe bugs that completely blocked
onboarding any tenant past the first (a hardcoded role-ID bug, a
global-instead-of-per-tenant first-time-setup gate, and two pre-existing
single-tenant `UNIQUE` indexes on `roles`/`templates` — migration
`v6.8.0`; see Decisions log); phase 7 (frontend audit) done — both
checklist items confirmed already satisfied by the existing
hard-navigation login/logout flow, no code changes needed; phase 9 not
started**. This document captures research and a phased
implementation plan for adding multi-tenancy to listmonk. It is an internal
engineering design doc, not end-user documentation.

See also: [`multi-tenancy-code-plan.md`](./multi-tenancy-code-plan.md) for the
concrete code-level plan (migration SQL, Go signatures, file-by-file changes)
and [`multi-tenancy-erd.drawio`](./multi-tenancy-erd.drawio) for entity-relationship
diagrams of the current and proposed schema (open at [app.diagrams.net](https://app.diagrams.net)
or the VS Code Draw.io Integration extension).

## Goal

Run a single listmonk instance (one binary, one database) serving multiple fully
isolated organizations ("tenants"). Each tenant gets its own users, lists,
subscribers, campaigns, templates, media, and settings (SMTP, from-address,
branding, etc.) — nothing is shared between tenants except the running process
and database server itself.

## Prior art

listmonk does not have multi-tenancy today, and upstream (`knadh/listmonk`) has
consistently declined to build it in:

- [#2872 — Native Multi-tenancy Support (Multiple Domains/Workspaces)](https://github.com/knadh/listmonk/issues/2872):
  the closest match to this plan — one instance serving multiple
  domains/brands with isolated SMTP config, from-addresses, archive pages, and
  subscriber data. **Closed as "not planned"** by maintainers, no rationale given.
- [#2765 — multi-tenancy / namespace feature](https://github.com/knadh/listmonk/issues/2765):
  proposes a `namespace` column on relevant tables, per-user namespace
  assignment, and `WHERE namespace = xxx` filtering on every query — i.e. the
  same row-level approach this doc lands on. Open, unaddressed.
- [#2395 — Multi-Tenancy Permissions](https://github.com/knadh/listmonk/issues/2395):
  narrower ask (a `lists:create` permission so non-admins can own their own
  lists). Closed as not planned.

**Implication:** this is a fork-only feature. It should be built and maintained
as a divergence from upstream, not something we can expect to upstream or stay
trivially rebasable against — every future `git merge` from `knadh/listmonk`
that touches `schema.sql`, `internal/core/*`, `internal/auth/*`, or
`internal/manager/*` will need conflict resolution against the tenant-scoping
changes.

## Architecture decision

- **Isolation model:** shared database, row-level isolation via a `tenant_id`
  column + PostgreSQL Row-Level Security (RLS), rather than schema-per-tenant
  or database-per-tenant. Lowest operational overhead (one connection pool,
  one set of migrations, one backup job), and RLS pushes isolation enforcement
  into the database instead of relying on every hand-written query in
  `internal/core/*` remembering a `WHERE tenant_id = ...` clause.
- **Tenant scope:** tenants are fully isolated organizations — separate users,
  lists, subscribers, campaigns, templates, media, and **settings** (SMTP
  creds, from-address, branding). Nothing is shared across tenants.

### Why RLS over app-level filtering

- [AWS: Multi-tenant data isolation with PostgreSQL Row Level Security](https://aws.amazon.com/blogs/database/multi-tenant-data-isolation-with-postgresql-row-level-security/)
- [AWS Prescriptive Guidance: Row-level security recommendations](https://docs.aws.amazon.com/prescriptive-guidance/latest/saas-multitenant-managed-postgresql/rls.html)

Key points from research:

- RLS filters rows inside Postgres itself — a missed `WHERE tenant_id = $1` in
  one of the ~40 hand-written queries in `queries/*.sql` can't leak data
  across tenants, because the policy is enforced even if the application query
  forgets it.
- The recommended pattern uses a **runtime session variable**, not a
  per-tenant Postgres role:
  ```sql
  ALTER TABLE subscribers ENABLE ROW LEVEL SECURITY;

  CREATE POLICY tenant_isolation_policy ON subscribers
    USING (tenant_id = current_setting('app.current_tenant')::UUID);
  ```
  The app sets `app.current_tenant` once per request/transaction (via
  `SET LOCAL` inside a transaction, so it auto-resets when the transaction
  ends) instead of creating a Postgres role per tenant.
- **Critical gotchas identified in research:**
  - RLS does **not** apply to superusers or roles with `BYPASSRLS`. The
    listmonk DB role used by the app must not have `BYPASSRLS` or superuser,
    or every policy is silently ignored.
  - `SET LOCAL app.current_tenant = ...` session variables are
    **incompatible with transaction-level poolers like PgBouncer** in
    transaction-pooling mode in some configurations — needs explicit testing
    against however listmonk pools connections (`sqlx` + `lib/pq`/`pgx`
    pool) before relying on it in production.
  - `tenant_id` must be the **leading column** in composite indexes on
    every scoped table, or RLS predicate filtering silently kills index
    usage (cited as "two orders of magnitude slower" without it).

## Current architecture (as of this branch)

Mapped against `schema.sql`, `internal/core/`, `internal/auth/`,
`cmd/handlers.go`, `cmd/init.go`, and `internal/manager/`:

### Tables needing a `tenant_id` column

| Table | PK today | Notes |
|---|---|---|
| `subscribers` | `id` SERIAL | unique `email` must become composite `(tenant_id, email)` |
| `lists` | `id` SERIAL | |
| `subscriber_lists` | `(subscriber_id, list_id)` | |
| `templates` | `id` SERIAL | unique `is_default` must become per-tenant |
| `campaigns` | `id` SERIAL | unique `archive_slug` must become composite |
| `campaign_lists` | `id` BIGSERIAL | |
| `campaign_views` | `id` BIGSERIAL | |
| `media` | `id` SERIAL | |
| `campaign_media` | (implicit) | |
| `links` | `id` SERIAL | |
| `link_clicks` | `id` BIGSERIAL | |
| `bounces` | `id` SERIAL | |
| `mat_dashboard_counts`, `mat_dashboard_charts`, `mat_list_subscriber_stats` | (matviews) | need a `tenant_id` grouping column and per-tenant (or filtered) refresh |
| `users` | `id` SERIAL | unique `username`/`email` must become composite; a super-admin concept that spans tenants needs explicit design (see Open questions) |

### Tables staying global

| Table | Notes |
|---|---|
| `roles` | scoped indirectly via `users.tenant_id`, but the row itself isn't tenant data today — needs `tenant_id` too since roles are per-org (`SuperAdminRoleID = 1` shortcut in `internal/auth/models.go` needs rethinking per tenant) |
| `sessions` | opaque `simplesessions` store; session payload already carries the resolved user, so tenant flows through the user, not the session row |
| **new:** `tenants` | new top-level table: `id`, `slug`/`domain`, `name`, `status`, plan/tier if needed |

### `settings` — decided: fully per-tenant, no split

`internal/core/settings.go` + `models/settings.go` currently model settings as
a single flat global key/value table (`settings.key TEXT UNIQUE`), pre-seeded
with ~50 keys (`app.*`, `smtp`, `bounce.*`, `privacy.*`, `security.*`,
`upload.*`, `appearance.*`).

**Decision:** every key becomes per-tenant, including `smtp`, `security.oidc`,
and `upload.s3.*` — each tenant configures its own SMTP server, OIDC identity
provider, and S3 bucket/credentials. There is no global/per-tenant split to
design: `settings` just gets a composite key `(tenant_id, key)` (already what
the phase 1 migration does) and `models.Settings` stays a single struct,
always loaded/saved scoped by `tenant_id` — no `TenantSettings` split type
needed. `Core.GetSettings`/`UpdateSettings` gain a `tenantID` parameter like
every other `Core` method (same pattern as phase 4's tenantID threading).

Two knock-on effects worth flagging for implementation:
- **OIDC callback URL becomes tenant-aware.** Per-tenant OIDC means the
  callback route (`/auth/oidc/callback` or similar) must resolve tenant from
  the request's subdomain *before* it can look up which IdP config to
  validate the auth-code exchange against — the callback handler needs the
  subdomain-resolution middleware (see below) to have already run.
- **S3 client construction becomes per-tenant.** `internal/media`'s S3 client
  is currently built once at startup from global settings. It needs to become
  lazily-constructed-per-tenant (or a per-tenant client cache keyed by
  `tenant_id`), mirroring the SMTP-pool-per-tenant change in phase 6.

### Request flow changes

- **Decided: tenant resolution is subdomain-based** —
  `<tenant-slug>.listmonk.example.com`. `cmd/init.go:initHTTPServer` has a
  global `srv.Use` that runs before auth on every request (currently just
  `c.Set("app", app)`) — this is the insertion point for a new tenant-
  resolution middleware: parse `c.Request().Host`, strip the port, extract
  the leftmost label, look up the tenant by `slug` (short-TTL in-memory cache,
  e.g. 30s, to avoid a DB round-trip per request), and `c.Set(TenantCtxKey,
  tenant)`. Unknown slug → 404. Tenant `status != active` → a generic
  "workspace unavailable" page (don't leak suspended-vs-disabled to the
  end user).
- `internal/auth/auth.go`: `Auth.Middleware` resolves a user from session or
  API token and does `c.Set(auth.UserHTTPCtxKey, user)`. After tenant
  resolution runs first (above), auth adds one check: the resolved user's
  `TenantID` must equal the middleware-resolved tenant's ID, or the request
  is rejected — defense in depth against a session/token issued on one
  tenant being replayed against another tenant's subdomain.
- **Session cookie scoping:** host-only cookies (no explicit `Domain`
  attribute), not `.listmonk.example.com`-wide. Since one user belongs to
  exactly one tenant, there's no login-once-use-everywhere requirement, and
  host-only cookies mean a stolen cookie can't even be replayed against a
  different subdomain — an extra isolation layer on top of the tenant-match
  check above.
- **DNS/TLS:** production needs a wildcard DNS record (`*.listmonk.example.com`)
  and a wildcard TLS cert (Let's Encrypt DNS-01 challenge — HTTP-01 doesn't
  support wildcards). For local dev, `*.localhost` resolves to `127.0.0.1`
  in modern browsers/OSes with no `/etc/hosts` edits — use e.g.
  `tenant-a.localhost:8080`.
- **Single-tenant/self-hosted fallback:** some self-hosters won't want
  subdomain routing for a single org. Add a config flag
  (`app.multi_tenancy_enabled`, default matching current behavior) that skips
  the tenant middleware entirely and pins every request to tenant id 1 (the
  default-tenant backfill from phase 1 already guarantees this works) — keeps
  a plain single-tenant upgrade painless.
- Every `internal/core/*` query needs a `tenantID` parameter threaded through,
  mirroring the existing `getAll bool, permittedIDs []int` pattern already
  used for list-permission scoping (e.g. `GetLists`, `QueryCampaigns`,
  `DeleteLists`) — this is the existing precedent to extend, not a new pattern
  to invent. With RLS in place this becomes a belt-and-suspenders check, not
  the sole enforcement mechanism.
- `internal/manager`: `Manager.scanCampaigns` → `Store.NextCampaigns` →
  `queries/campaigns.sql:next-campaigns` currently pulls due campaigns
  globally with no tenant filter. Needs either a tenant-aware query per
  worker cycle, or the worker pool needs to iterate tenants and set
  `app.current_tenant` per batch. Per-tenant rate limiting
  (`app.message_rate`/`app.concurrency`, currently global settings) also
  needs a design decision — global cap shared across tenants vs. per-tenant caps.

## Operator API (cross-tenant management)

**Decided:** cross-tenant actions (suspend/reactivate a tenant, billing hooks)
are exposed as a dedicated REST API, not the per-tenant admin UI — there is no
UI-level "operator" role.

- **Auth:** a static bearer token from config (`LISTMONK_operator__token` /
  `[operator] token=` in `config.toml`), checked by a dedicated Echo
  middleware on its own route group (e.g. `/api/operator/*`), entirely
  independent of the session/JWT auth used by tenant users. Rotated by
  changing config + restart — no DB-backed revocation in v1.
- **DB access:** operator routes run through a distinct Postgres role with
  `BYPASSRLS` — this must **not** be the same role the tenant-facing app pool
  uses (see the RLS gotcha above: `BYPASSRLS`/superuser silently disables every
  policy). This role becomes the second consumer of the "separate maintenance
  role" already flagged for migrations/backups in the phased plan below.
- **Endpoints (v1, minimal):**
  - `GET /api/operator/tenants` — list all tenants + status + basic counts
    (users, subscribers) for a support/billing dashboard.
  - `GET /api/operator/tenants/:id` — tenant detail.
  - `POST /api/operator/tenants` — provision a new tenant (slug, name, initial
    admin user).
  - `PUT /api/operator/tenants/:id/status` — suspend/reactivate/disable
    (`{"status":"suspended"}`); a suspended tenant's subdomain shows the
    "workspace unavailable" page (from tenant-resolution middleware above),
    and the manager/dispatcher scan (phase 6) skips it.
- **Out of scope for this plan:** actual payment-provider integration
  (Stripe subscription state, invoicing). The operator API just provides the
  status/suspend lever a billing webhook handler would call into.

## Phased implementation plan

1. **Schema foundation**
   - New `internal/migrations/vX.Y.Z.go` migration: create `tenants` table;
     add `tenant_id` (nullable initially) to every table listed above; backfill
     a single default tenant for existing installs so upgrades are non-breaking.
   - Update `schema.sql` for fresh installs.
2. **RLS policies**
   - Enable RLS on every tenant-scoped table; add `USING (tenant_id =
     current_setting('app.current_tenant')::INT)` policies.
   - Confirm the app's Postgres role has no `BYPASSRLS`/superuser grant.
   - Add composite indexes with `tenant_id` leading on every scoped table.
3. **Connection/session plumbing**
   - Verify how `sqlx`'s underlying driver and connection pool interact with
     `SET LOCAL` — spike a small test to confirm tenant context isn't leaked
     across reused pooled connections before writing any handler code.
   - Add a helper (e.g. `Core.WithTenant(ctx, tenantID)`) that wraps a
     transaction, issues `SET LOCAL app.current_tenant`, and runs the query.
4. **Auth & request flow**
   - Add `TenantID` to `auth.User`/session payload.
   - Insert the subdomain tenant-resolution middleware into the request path
     (see "Request flow changes" above), running before auth; auth then
     cross-checks the resolved user's `TenantID` matches.
   - Migrate `users`/`roles` uniqueness constraints to be tenant-composite.
5. **Settings (fully per-tenant)**
   - Migrate `settings` to `(tenant_id, key)` composite key (no split logic —
     every key, including `smtp`/`security.oidc`/`upload.s3.*`, is per-tenant);
     update `models.Settings` and `Core.GetSettings`/`UpdateSettings` to take
     `tenantID`. Note the OIDC-callback and S3-client knock-on effects above.
6. **Manager/dispatcher**
   - Make `NextCampaigns` and the worker scan loop tenant-aware; decide on
     per-tenant vs. global send-rate limiting; skip tenants with
     `status != active`.
7. **Frontend**
   - Tenant switcher (if a user can belong to only one tenant, this may be
     unnecessary); ensure no cross-tenant IDs leak through the Pinia store or
     orval-generated API client responses.
8. **Public-facing routes**
   - Subscription pages, unsubscribe links, campaign archive pages, and
     tracking pixels/link-click redirects are all public/unauthenticated.
     Since tenant is now resolved from the subdomain before any handler runs
     (not from the entity's own row), the audit is: confirm every lookup in
     these handlers filters by the subdomain-resolved `tenant_id`, and that a
     valid UUID from tenant A can't be used to fetch data while sitting on
     tenant B's subdomain.
9. **Operator API**
   - New route group + static-token middleware + `BYPASSRLS` operator DB role
     (see "Operator API" above); tenant list/detail/provision/suspend
     endpoints.

## Decisions log

- **Tenant resolution:** subdomain (`<slug>.listmonk.example.com`), resolved
  by middleware before auth. See "Request flow changes" above.
- **Cross-tenant management:** no UI-level super-admin. A separate Operator
  API (static bearer token, `BYPASSRLS` DB role) handles tenant
  provisioning/suspension/billing hooks. See "Operator API" above.
- **Settings:** no global/per-tenant split — every key, including
  `smtp`/`security.oidc`/`upload.s3.*`, is per-tenant. See "`settings` —
  decided" above.
- **Phase 1 implementation (2026-07-07):** shipped purely additive — every
  scoped/join table + `settings` gets `tenant_id INTEGER NOT NULL DEFAULT 1
  REFERENCES tenants(id)`, no existing constraint touched. The original
  code-plan draft prematurely re-scoped uniqueness constraints
  (`subscribers.email`, `users.username`/`email`, etc.) in the same
  migration — that would have broken `ON CONFLICT` upserts in
  `queries/*.sql` ahead of the query-layer changes. Corrected; see
  `multi-tenancy-code-plan.md`'s Phase 1 section for the full explanation.
- **Phase 2 implementation (2026-07-07):** RLS enabled + forced on all 15
  tables, with a deliberately permissive policy (`... OR
  current_setting(...) IS NULL`) rather than the original draft's strict
  "unset context sees nothing" — the strict version would break every
  query the moment the migration runs on a correctly-permissioned
  deployment, since the app doesn't set `app.current_tenant` until phase 4.
  Also added `FORCE ROW LEVEL SECURITY` (owners are RLS-exempt by default,
  and most self-hosted installs use one Postgres role for everything).
  Verified with a throwaway non-superuser role since this dev DB's own role
  is a superuser and would bypass RLS regardless of the policy. **Tighten
  the permissive fallback once phase 4 lands** — tracked as a follow-up on
  issue #29, not a new phase. See `multi-tenancy-code-plan.md`'s Phase 2
  section for detail.
- **Phase 3 implementation (2026-07-07):** `Core.WithTenant` (a method, not
  the originally-drafted package-level function) plus a permanent
  concurrency test — this repo's first Go test, since `go test ./...` had
  zero test files before this. The test creates its own throwaway
  least-privileged Postgres role rather than using the app's configured
  role, for the same reason phase 2's manual verification needed one: both
  this dev DB's role and CI's default Postgres image role are superusers,
  which bypass RLS regardless of policy. See `multi-tenancy-code-plan.md`'s
  Phase 3 section for detail, including a `t.Cleanup`-ordering bug caught
  and fixed along the way.
- **Phase 4 implementation (2026-07-07) — scope split:** investigation
  found threading `tenantID` through `internal/core` (the other half of
  the originally-drafted phase 4) touches 107 exported methods and 150
  call sites, none of which take `context.Context` today or run through a
  transaction — too large to do safely alongside the auth/resolution work.
  Split into its own follow-up issue; this session shipped only subdomain
  tenant resolution, `TenantID` on `auth.User`, and the auth cross-check,
  gated behind a new `app.multi_tenancy_enabled` config flag (default
  `false`) so it's additive like phases 1-3. Also discovered and reused:
  the `Tenant` struct and shared context-key constant had to live in
  `models` rather than a new `internal/tenant`-only type, to avoid an
  import cycle (`internal/core` already imports `internal/auth`). Verified
  end-to-end (login on one tenant's subdomain, confirmed session rejected
  when replayed against a different tenant's subdomain) via a temporary
  second backend instance, not the live dev one. See
  `multi-tenancy-code-plan.md`'s Phase 4 section for full detail.
- **Issue #40 slice 1 (`subscribers.go`) implementation (2026-07-07):**
  first slice of the deferred `internal/core` tenantID-threading work.
  Found and fixed a critical, repo-wide sqlx footgun that will matter for
  every future slice: `sqlx.Tx.Stmtx()` silently drops the `.Unsafe()`
  flag the pool `*sqlx.DB` is opened with (needed since every table has an
  unmapped `tenant_id` column post-phase-1) — fixed via a `stmtx()` helper
  in `internal/core/tenant.go` that all future slices must use instead of
  calling `tx.Stmtx()` directly. Verified real cross-tenant isolation at
  the RLS/SQL level (via a non-superuser role, same technique as phase
  2/3) since this dev DB's superuser role can't demonstrate it end-to-end
  over HTTP — a pre-existing limitation, not a defect in this slice. See
  `multi-tenancy-code-plan.md`'s new "Issue #40" section for full detail.
- **Phase 5 implementation (2026-07-07) — scope split:** settings aren't
  read per-request — they're loaded once at process boot into a global
  config, and SMTP pools/media store/OIDC config/campaign manager are all
  built once as process-lifetime singletons from it (no live-reload
  mechanism exists today even for the current single-tenant flow — it's a
  full `syscall.Exec` process restart). Making those subsystems genuinely
  per-tenant is a redesign, not a parameter addition — split into issue
  #41. Shipped: the DB/`Core` layer (migration `v6.6.0`'s composite
  `(tenant_id, key)` key, `Core.GetSettings`/`UpdateSettings`/
  `UpdateSettingsByKey` now tenant-scoped), with those four subsystems
  left as global singletons pinned to tenant 1 (documented in code).
  Found and fixed a real bug while *running* the migration (not caught by
  reading code alone): the upgrade runner's own version-bookkeeping query
  in `cmd/install.go` depended on the constraint this migration removed —
  the same class of issue phase 1 deferred constraint changes to avoid,
  slipping through here because it lives in the migration framework
  itself, not `queries/*.sql`. See `multi-tenancy-code-plan.md`'s Phase 5
  section for full detail.
- **Phase 6 / issue #41 slice 1 (2026-07-07):** confirmed Phase 6
  (tenant-aware campaign scanning) is only meaningful once real per-tenant
  dispatch exists — chose to tackle #41 (the actual blocker) directly
  rather than do Phase 6's scan-side work in isolation. Shipped #41's
  first slice: per-tenant SMTP messenger resolution. `internal/manager`
  gained a `MessengerResolver` interface (falls back to the existing
  global messenger map for non-SMTP messengers like postback, so nothing
  else changes); the concrete implementation lives in `cmd/` and lazily
  builds + caches each tenant's SMTP messengers from their own settings,
  reusing (not reimplementing) the existing SMTP-config-parsing logic.
  Verified live by starting a real campaign and confirming the resolution
  path executes (visible as a second `initSMTPMessengers` log line at
  campaign-start, not just boot) before failing at the expected point — an
  actual SMTP connection attempt to this dev environment's fake mail host.
  Media/OIDC per-tenant resolution remain as further #41 work.

- **Phase 6 implementation (2026-07-08):** unblocked by #41 slice 1,
  implemented the same session. `Store.NextCampaigns` gained a leading
  `tenantID` param and a new `GetActiveTenantIDs` method;
  `scanCampaigns` iterates active tenants each tick, calling
  `NextCampaigns` once per tenant with only that tenant's in-flight
  campaign IDs. Traced through a real correctness risk *before* writing
  code: `next-campaigns` reuses its "current IDs" parameter to both
  exclude in-flight campaigns *and* increment their `sent` counts, so
  naively passing the same global in-flight list to every per-tenant call
  would double/triple/etc.-count sends — fixed by grouping in-flight
  campaigns by tenant in Go, not by changing the SQL's counting logic.
  Separately, **caught a real bug via live-testing, not code review**:
  `pq.Int64Array(nil)` serializes to SQL `NULL` rather than an empty
  array, and `NOT(id = ANY(NULL))` is `NULL` (falsy) under normal SQL
  three-valued logic — silently filtering out every campaign, no error
  anywhere. A `running` campaign sat unpicked-up for several tick cycles
  before this was caught by directly observing the symptom, then
  root-caused by comparing against the original single-tenant code's
  explicit non-nil-empty-slice construction (which its own comment had
  flagged the reason for). Verified live end-to-end (started a real
  campaign, confirmed pickup on the next tick, confirmed the SMTP
  resolver and eventual expected-failure send all fired correctly) both
  before and after the fix, to prove the bug and the fix were both real.

- **Cross-cutting INSERT/RLS gap found and fixed (2026-07-08):** while
  starting issue #40 slice 2 (`internal/core/campaigns.go`), found that
  Phase 2's RLS policy (`v6.5.0`) has no explicit `FOR`/`WITH CHECK`
  clause, so Postgres defaults an `ALL`-command policy's `WITH CHECK` to
  the same expression as `USING` — meaning **every `INSERT`** across the
  whole schema must supply a `tenant_id` value matching
  `current_setting('app.current_tenant')`, or a non-superuser DB role
  rejects the row. Checked every `queries/*.sql` file: **zero** `INSERT`
  statements set `tenant_id` explicitly (they all relied on the column's
  `DEFAULT 1` from Phase 1) — meaning every write for any tenant other
  than tenant 1 would have been rejected under real (non-superuser) RLS
  enforcement. Invisible until now because the dev DB role is a
  superuser (RLS-exempt entirely — same reason Phase 2/3 needed a
  throwaway non-superuser role to demonstrate isolation at all).
  User chose the broad fix (over scoping to just campaigns.go or
  deferring): swept every `INSERT` across `campaigns.sql`, `links.sql`,
  `subscribers.sql`, `bounces.sql`, `lists.sql`, `media.sql`, `roles.sql`,
  `templates.sql`, `users.sql`, adding an explicit `tenant_id` param to
  each. This required threading `ctx`/`tenantID` through essentially all
  of `internal/core` in the process — `campaigns.go`, `subscriptions.go`,
  `bounces.go`, `lists.go`, `media.go`, `roles.go`, `templates.go`,
  `users.go` — which amounts to completing the large majority of issue
  #40 (Core-threading) as a side effect, not just its originally-planned
  slice 2. Also fixed two call paths that bypass Core entirely and thus
  never ran through `WithTenant`: the bulk CSV importer
  (`internal/subimporter`, gained a `SessionOpt.TenantID` set by the
  HTTP handler from the resolved request tenant) and
  `cmd/manager_store.go`'s `CreateLink`/`GetAttachment`/
  `GetInlineAttachmentByFilename` (called from `internal/manager`'s
  campaign-send pipeline, now threaded via `models.Campaign.TenantID`
  down through `Manager.trackLink`/`attachMedia`/`applyInlineImages`).
  **Also found and fixed a related, adjacent bug in the same file**:
  `LoginUser`'s query had zero tenant filtering — since `username` is
  still a global `UNIQUE` constraint (Phase 1, deferred), a valid
  username/password for tenant A would also log in successfully on
  tenant B's subdomain if a same-named account existed there. Fixed by
  routing `LoginUser` through `WithTenant` like everything else, so RLS
  narrows the match to the resolved tenant. This is distinct from (and
  was previously masked by) `internal/auth.tenantMismatch`, which only
  guards *replaying* an existing session/token against the wrong tenant,
  not the login call itself.
  Two auth-adjacent lookups were deliberately kept **unscoped** (new
  `Core.GetUserUnscoped`/`GetAllUsersUnscoped`, with comments explaining
  why): the session/API-token lookup callback
  (`cmd/init.go`'s `auth.Callbacks.GetUser`) must find a user regardless
  of tenant so `tenantMismatch` can compare and reject with a distinct
  403 rather than RLS silently making it look like "not found"; and the
  API-token cache (`cmd/users.go`'s `cacheUsers`) must hold every
  tenant's API users since an incoming token's tenant is only known
  after the token itself is matched.
  Found one additional bug via live-testing that pure code review had
  missed: `CreateRole`/`CreateListRole` were written accepting
  `ctx`/`tenantID` params but forgot to actually pass `tenantID` as the
  new trailing SQL arg — caught immediately via a live `curl` against the
  dev server (`sql: expected 4 arguments, got 3`), not by `go build`
  (prepared-statement `.Exec`/`.Get` take variadic `any` args, so
  arg-count mismatches are runtime-only). This reinforced the
  session-wide pattern that live verification catches classes of bugs
  code review and `go build`/`go vet` cannot. Verified live end-to-end
  after the fix: created a list, subscriber (with list subscription),
  template, user role, list role (with list permissions), user, media
  upload, and a campaign (with list + media attachment), then ran the
  campaign to completion (draft → running → finished) confirming
  `campaign_media`, `campaign_lists`, and `links` (via a real
  `@TrackLink`-tagged URL, exercising the previously-Core-bypassing
  `CreateLink` path) all wrote rows with the correct `tenant_id` before
  the expected fake-SMTP-host send failure. All test data cleaned up
  afterward. No new migration needed — `tenant_id` columns already
  existed from Phase 1; this was purely a `queries/*.sql` +
  `internal/core`/`cmd` fix.

- **Matview cross-tenant data leak found and fixed, closing issue #40
  (2026-07-08):** picking up the one remaining #40 file
  (`internal/core/dashboard.go`) surfaced that this was more than a
  threading gap. `mat_dashboard_counts`, `mat_dashboard_charts`, and
  `mat_list_subscriber_stats` (this doc's own "Matview refresh cost"
  open question, until now) each computed a **single global row** with
  no `tenant_id` column at all. Worse than the write-rejection shape of
  the INSERT gap above: this was a **live cross-tenant read leak**.
  `queries/subscribers.sql`'s `query-subscribers-count-all` falls back to
  `mat_list_subscriber_stats`'s `list_id=0` "all subscribers" row
  whenever a request has no list filter — the common case (Subscribers
  page load, no search) — so every tenant's unfiltered subscriber total
  was silently the sum across every tenant's subscribers, not just their
  own. Same shape for the dashboard's subscriber/list/campaign totals and
  the click/view charts. Given the severity, asked the user whether to
  fix now, file an issue, or explain first — user chose **fix now**.
  Migration `v6.7.0` rewrites all three materialized views to compute
  one row per tenant (driven by a join/group against the `tenants`
  table), widens each view's required unique index to lead with
  `tenant_id` (needed for `REFRESH MATERIALIZED VIEW CONCURRENTLY`,
  already used by `Core.RefreshMatView`), and adds an explicit
  `tenant_id` filter to `get-dashboard-charts`/`get-dashboard-counts`/
  `query-subscribers-count-all`. The refresh mechanism itself is
  unchanged — one `REFRESH` statement still refreshes every tenant's row
  at once, matching this doc's own recommended default (keep the global
  refresh cadence, filter by `tenant_id` at query time) instead of
  building per-tenant incremental refresh. `dashboard.go`'s two methods
  now take `ctx`/`tenantID` like the rest of `internal/core`, closing out
  issue #40 completely.
  **Found a second instance of the exact same class of bug already fixed
  once this session**: `GetDashboardCharts`/`GetDashboardCounts` were
  written accepting `ctx`/`tenantID` and correctly opening a
  `WithTenant` transaction, but the actual `.Get(&out, ...)` call forgot
  to pass `tenantID` as the query's new `$1` arg — caught immediately via
  live `curl` (`sql: expected 1 arguments, got 0`), not by `go build`/`go
  vet`. Fixed and reverified live. Verified end-to-end against the dev
  DB: ran the migration, confirmed via `psql` that each matview now has
  one correctly-scoped row per tenant, then hit `/api/dashboard/counts`,
  `/api/dashboard/charts`, and `/api/subscribers` over HTTP and confirmed
  correct tenant-1-only numbers.

- **Issue #41 slice 2 — media/S3 per-tenant store resolution implemented
  (2026-07-08):** same shape as #41 slice 1 (SMTP): `upload.*` settings
  were already per-tenant (Phase 5), but the `media.Store` consumer was
  still a single process-lifetime singleton built once at boot. New
  `cmd/tenant_media.go`'s `tenantMedia` lazily builds and caches a
  `media.Store` per tenant from that tenant's own settings, reusing
  `initMediaStore` (refactored to return an error instead of calling
  `lo.Fatalf`, mirroring `initSMTPMessengers`'s Phase-41-slice-1 change,
  so a lazy per-tenant failure doesn't crash the whole process). Unlike
  the SMTP resolver, **no fallback map was needed** — every media
  consumer (`cmd/media.go`'s HTTP handlers, `cmd/manager_store.go`'s
  campaign-attachment methods on the real send path) already has a
  `tenantID` in scope, so `App.media`'s type changed outright from
  `media.Store` to `*tenantMedia`. Also fixed `cmd/admin.go`'s
  `GetServerConfig` (the frontend's config-bootstrap endpoint), which had
  the same latent bug as everywhere else pre-fix: it reported the global
  boot-time upload provider rather than the requesting tenant's actual
  one. Verified live: uploaded a file, confirmed `/api/config` reports
  the correct per-tenant provider, then ran a real campaign with that
  media attached to completion, confirming the campaign-send path's
  `GetAttachment` resolved correctly through the new per-tenant store
  before the expected fake-SMTP-host failure. Remaining in #41: only
  OIDC per-tenant config resolution.

- **Issue #41 slice 3 — OIDC per-tenant config resolution implemented,
  closes #41 (2026-07-08):** structurally different from the SMTP/media
  slices since `internal/auth` correctly never imports `internal/core` —
  extended the package's existing `Callbacks` bridge pattern
  (`GetCookie`/`SetCookie`/`GetUser`) with a new `GetOIDCConfig(tenantID)
  (OIDCConfig, error)` callback rather than introducing a `cmd/`-side
  resolver type. `Auth`'s single global `provider`/`verifier`/`oauthCfg`
  fields became a `map[int]*tenantOIDC` cache (reusing the struct's
  existing mutex). Unlike SMTP/media, there was no boot-time fail-fast to
  preserve — OIDC was already lazily initialized on first use even in the
  original single-tenant code, so no `initX(ko) (X, error)`-style refactor
  was needed. `RedirectURL` is computed in the callback (`cmd/init.go`)
  from that tenant's own `AppRootURL`, not a single global one, so the
  OAuth callback lands back on the correct subdomain per tenant.
  **A genuine design decision, not just mechanical threading**:
  `cmd/handlers.go` only registered `/auth/oidc` at boot if the global
  config had OIDC enabled — under true per-tenant OIDC, one tenant could
  enable it while the boot-time snapshot (tenant 1's settings) doesn't,
  or vice versa. Resolved by registering the routes when
  `app.multi_tenancy_enabled` is true (request-time per-tenant checks
  take over from there) **or** the existing global flag is true
  (unchanged behavior for today's default single-tenant deployments).
  Also fixed the same "global config leaking into a per-tenant response"
  bug class as slice 2's `GetServerConfig`, this time in
  `renderLoginPage` (which OIDC button to show) and `createOIDCUser`
  (which default role IDs to auto-provision with) — both now read the
  resolved tenant's `GetSettings` instead of `a.cfg.Security.OIDC.*`.
  **Verified live against a real IdP**, not a stub: enabled OIDC via the
  Settings API pointed at `https://accounts.google.com` with a fake
  client ID, confirmed the settings-restart flipped `/auth/oidc` from 404
  to registered, then confirmed a real `302` redirect to a genuine Google
  OAuth URL with the correct `client_id` and a `redirect_uri` derived
  from tenant 1's own root URL — proving `oidc.NewProvider`'s real network
  discovery succeeded and the whole resolver chain was live, not
  hypothetical. Restored OIDC to disabled and confirmed the route
  reverted to 404 afterward. Issue #41 is now fully complete.

- **Phase 8 — public-facing route audit implemented (2026-07-08): no
  cross-tenant leak in public routes, but 3 severe onboarding-blocking
  bugs found and fixed.** Audited every multi-UUID public route
  (`ViewCampaignMessage`, `LinkRedirect`, `RegisterCampaignView`,
  `SubscriptionPrefs`) against this doc's own audit criterion — does a
  second UUID/ID param get trusted without re-checking it belongs to the
  same resolved tenant as the first? Traced every underlying SQL
  statement rather than assuming: all of them run inside the same
  `WithTenant` transaction, so RLS scopes every table reference to the
  same `app.current_tenant` uniformly — a cross-tenant UUID inside a
  compound query resolves to zero rows for that sub-lookup, either
  degrading to a mis-attributed (not cross-tenant-leaked) record where
  the column is nullable, or hitting a pre-existing `NOT NULL`-triggered
  silent no-op where it isn't. **No cross-tenant data exposure found.**
  Live-testing the onboarding path itself, however (a real second tenant
  created via a throwaway DB row, tested via `curl -H "Host: ..."`
  overrides against `127.0.0.1:9000` — confirmed
  `internal/tenant.Middleware` reads `c.Request().Host` directly, so no
  `/etc/hosts` changes were needed), surfaced three real, severe bugs
  that blocked onboarding any tenant past the first entirely:
  1. `doFirstTimeSetup` hardcoded the new admin's `user_role_id` to the
     `SuperAdminRoleID` constant (1) instead of the ID of the role it had
     just created. Only ever worked for the very first tenant on an
     installation (whose first-ever `roles` row coincidentally gets
     id=1, since the sequence is shared across every tenant) — every
     subsequent tenant's new admin ended up with `user_role_id = NULL`
     (the RLS-scoped role lookup silently finds nothing for a role
     belonging to a different tenant), a login-successful but completely
     permission-less account. Fixed by capturing and using the actual
     created role's ID.
  2. `App.needsUserSetup` was a single process-wide boolean, computed
     once at boot from whether *any* tenant had *any* user and flipped
     `false` globally the first time *any* tenant completed setup — so
     every tenant after the first could never even reach the
     first-time-setup form via `/admin/login`. Fixed by removing the
     cached field and adding `Core.HasUsers(ctx, tenantID)` (a cheap
     `EXISTS` query), checked per-request, per-tenant.
  3. Two pre-existing single-tenant `UNIQUE` indexes — `roles(type,
     name)` and `templates(is_default) WHERE is_default = true` — had no
     `tenant_id` dimension and hard-blocked (not just weakened isolation
     for) legitimate usage: the *second* tenant to ever run setup
     hard-fails creating its "Super Admin" role (duplicate key), and no
     tenant but the first could ever set their own default template.
     Both predate phase 1's `tenant_id` columns and were deliberately
     left untouched by phase 1 (see that phase's decision to defer
     pre-existing constraints) — this is that deferred cleanup, scoped
     to the two that actively block usage. New migration `v6.8.0` widens
     both to `(tenant_id, ...)`. Two *other* known global-uniqueness
     gaps (`subscribers.email`, `links.url`, already flagged earlier
     this session) were deliberately left alone — they cause soft
     collisions on human-chosen values, not hard failures, and widening
     them changes `ON CONFLICT` semantics existing queries rely on.
  All three were surfaced to the user via `AskUserQuestion` given their
  severity before fixing — "fix now" chosen each time. Verified via a
  direct SQL/RLS simulation under a throwaway non-superuser role
  reproducing `doFirstTimeSetup`'s exact operations (confirmed the *old*
  code's failure reproduces exactly as diagnosed, then confirmed the
  fixed code succeeds with a correctly non-null `user_role_id`) — a full
  HTTP-level test wasn't possible since the dev DB connection is itself a
  superuser (RLS-exempt, the same limitation documented since phases
  2/3). The `v6.8.0` migration itself was verified directly against the
  dev DB via `psql`, and existing single-tenant login/template/role flows
  were confirmed still working normally afterward. All test artifacts
  (throwaway DB roles, temp tenant row, temporary `config.toml`
  overrides) cleaned up and reverted to baseline.

- **Phase 7 — frontend audit implemented (2026-07-08): both checklist
  items confirmed already satisfied, no code changes needed.** Verified
  by inspection rather than assumed: `frontend/src/App.vue`'s
  `doLogout()` does a hard `document.location.href` navigation (not a
  client-side route change), unconditionally destroying all in-memory
  Pinia state; confirmed no persistence plugin or `localStorage`/
  `sessionStorage` usage anywhere in the store that could survive it.
  Login is symmetric — server-rendered (`cmd/auth.go`'s `LoginPage`, a Go
  template) and completes via an HTTP redirect into a freshly-initialized
  SPA, so no prior session's state can carry over. Grepped the entire
  frontend for "tenant" — zero matches — confirming tenant scoping is
  fully server-side and transparent to the client, as designed.
  Additionally, under the subdomain-per-tenant model, two tenants' UIs
  are different browser origins by construction, so browser storage is
  origin-partitioned regardless — cross-*tenant* leakage isn't reachable
  even in principle without a deliberate architecture change. The
  original checklist concern (stale state surviving a user switch) was
  actually about same-tenant user hygiene, and that's already handled by
  the existing hard-navigation logout.

## Open questions

- **Matview refresh cost — resolved 2026-07-08:** `mat_dashboard_counts`,
  `mat_dashboard_charts`, and `mat_list_subscriber_stats` now carry a
  `tenant_id` dimension (migration `v6.7.0`), one row per tenant, and
  still refresh globally in one `REFRESH MATERIALIZED VIEW CONCURRENTLY`
  statement per `Core.RefreshMatViews` — the "recommended default" this
  entry originally proposed. What escalated this from a performance
  question to an actual fix: it turned out to be a **live cross-tenant
  data leak**, not just a future refresh-cost concern — see the Decisions
  log entry for what was actually leaking and how it was found. Move to
  per-tenant incremental refresh only if refresh time becomes a measured
  problem at scale.
- **Upgrade path for existing single-tenant installs:** the default-tenant
  backfill in step 1 needs to guarantee zero-downtime, reversible migration
  per this repo's [migration conventions](/CLAUDE.md) (idempotent, updates
  `schema.sql`, registered in `cmd/upgrade.go`'s `migList`). The
  `app.multi_tenancy_enabled` flag (see "Request flow changes") is the
  intended escape hatch for self-hosters who don't want subdomain routing.
- **Operator provisioning UX:** `POST /api/operator/tenants` creates a tenant
  + initial admin user — needs a decision on how that admin's initial
  password/invite is communicated (return a one-time setup link in the API
  response? require the operator to set it out of band?).

## References

- [AWS: Multi-tenant data isolation with PostgreSQL Row Level Security](https://aws.amazon.com/blogs/database/multi-tenant-data-isolation-with-postgresql-row-level-security/)
- [AWS Prescriptive Guidance: Row-level security recommendations](https://docs.aws.amazon.com/prescriptive-guidance/latest/saas-multitenant-managed-postgresql/rls.html)
- [knadh/listmonk#2872 — Native Multi-tenancy Support](https://github.com/knadh/listmonk/issues/2872) (closed, not planned)
- [knadh/listmonk#2765 — multi-tenancy / namespace feature](https://github.com/knadh/listmonk/issues/2765) (open)
- [knadh/listmonk#2395 — Multi-Tenancy Permissions](https://github.com/knadh/listmonk/issues/2395) (closed, not planned)
