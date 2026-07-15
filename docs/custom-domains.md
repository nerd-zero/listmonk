# Custom domains for an instance — design

Status: **planned, not started**. This is `docs/plan.md`'s "v1.1 — bring-your-own domain" (build order step 6), broken out into its own doc because it touches both repos (listnun and the listmonk fork) and needs its own data model, API surface, and step-by-step flow the way `plan.md`'s per-feature sections already do.

## What this is (and isn't)

"Custom domain" is ambiguous in this app because there are two genuinely different things it could mean:

1. **Where the instance itself is reached** — the tenant's admin login, public unsubscribe/preference pages, campaign-in-browser view, and click-tracking links. Today this is always `{slug}.{root_domain}` (e.g. `acme.listnun.app`), set once at tenant creation and never changed. **This is what this doc is about.**
2. **What address campaigns send from** — already built. `AddSenderDomain`/`AddPlatformDomain` (`internal/provisioning/service.go`) let an org verify their own sending domain via DKIM + Return-Path, independent of what this doc adds.

Both are "a domain the org owns, pointed at their instance" in spirit, but they're unrelated pieces of infrastructure (SMTP sending vs. HTTP routing) with no shared code today, and should stay that way — conflating them would mean a DNS mistake on one breaks the other.

## Current state

- **listmonk fork**: `internal/tenant/resolve.go`'s `Middleware` resolves a tenant purely by stripping `.{root_domain}` off the request's `Host` header (`strings.TrimSuffix(host, "."+rootDomain)`) and looking up that slug. Any `Host` that isn't `<slug>.{root_domain}` 404s immediately — there's no path for an arbitrary domain to resolve to a tenant at all today.
- **listmonk fork**: `cmd/operator.go`'s `tenantRootURL` computes `app.root_url` once, at `CreateTenant` time, from the tenant's slug and `app.root_domain`. `SetTenantRootURL` (the query behind it) is only ever called from `CreateTenant` — there's no operator endpoint to change a tenant's `root_url` afterward.
- **listnun**: `internal/config.Config.ListmonkRootDomain` exists only to *display* a tenant's expected URL before its setup link is known — cosmetic, not used for routing.
- **Production is already behind Cloudflare** (confirmed 2026-07-15), which is what makes this tractable without standing up new TLS infrastructure ourselves: **Cloudflare for SaaS** (Custom Hostnames API) issues and renews a cert per customer domain and proxies matching traffic to a fallback origin we configure — already researched in `docs/plan.md`'s "Custom domains without exposing an IP" section. Nothing about that research is stale; only the "no Cloudflare Tunnel exists yet" framing needs re-checking against however production is actually fronted today (see Open decisions).
- **The DNS-verification UX already exists and is exactly the shape this needs**: `sender_identities`/`dns_records` (`db/migrations/0001_init.up.sql`, `0002_sender_identities.up.sql`), `GetSenderIdentity`'s re-check-on-read pattern (calls Postmark live and persists `pending` → `confirmed` the next time the identity is fetched, rather than needing a background poller), and `SenderIdentityCard`'s per-field copy buttons (`web/src/components/sender-identity-card.tsx`). This doc reuses that pattern rather than inventing a new one.

## Proposed flow

1. **Org enters a domain** (e.g. `mail.acme.com`) for one of their instances, in the UI.
2. **Backend registers a Cloudflare Custom Hostname** for it — `POST /custom_hostnames` on our zone, fallback origin = wherever the shared listmonk deployment is actually reachable from Cloudflare's edge (see Open decisions). Response includes the ownership-verification record Cloudflare requires before it'll issue a cert.
3. **UI shows the DNS records to publish** — the CNAME (`mail.acme.com` → our Custom Hostname's target) and Cloudflare's ownership-verification TXT — using the same record-list-with-copy-buttons component pattern as `SenderIdentityCard`.
4. **Org publishes the records**, then either we poll or they click a "Verify" action that re-checks Cloudflare's Custom Hostname status live (`GET /custom_hostnames/{id}`) and persists `pending` → `active` the moment `ssl.status == "active"` — same on-read-recheck pattern as `GetSenderIdentity`, no new polling infrastructure needed for v1.
5. **Once active, listnun calls a new operator endpoint** (needs building — see below) to update the tenant's `app.root_url` to the custom domain.
6. **The org's domain is live.** Their admin login, public pages, and tracking links now resolve under their own domain; `{slug}.{root_domain}` presumably keeps working too (see Open decisions — whether the default subdomain stays reachable after a custom domain is set).

Removing a custom domain reverses steps 5 and 2: revert `app.root_url` to `{slug}.{root_domain}`, then delete the Cloudflare Custom Hostname.

## Work required

### listmonk fork

- **New operator endpoint**: `PUT /api/operator/tenants/{id}/root-url` (naming to match the existing `.../status` and `.../smtp` siblings in `cmd/operator.go`), backed by the already-existing `SetTenantRootURL` query — it just needs an HTTP handler and operator-store method calling it outside of `CreateTenant`.
- **`internal/tenant/resolve.go`'s `Middleware` needs a second resolution path**: today it only ever strips `.{root_domain}`. It needs to also look up a tenant by an exact `Host` match against a new `custom_domain`-type column (or a small side table) before falling back to (or instead of) subdomain stripping. This is the one piece of this feature that's a genuine behavior change to the tenant-resolution hot path, not just new plumbing alongside it — worth its own careful review and test pass in that repo, mirroring the rigor `docs/design/multi-tenancy.md`'s own phases already applied to the subdomain path.
- **Schema**: `tenants` needs the new column (or a `tenant_custom_domains` table, if it should support the domain being pending/unverified independent of `app.root_url` actually being updated yet — recommend this shape, so the fork doesn't resolve a domain to a tenant until listnun has confirmed it and flipped `root_url`).

### listnun

- **New Cloudflare client** (`internal/cloudflareclient`), thin hand-rolled HTTP client — same "no SDK, small surface" reasoning already applied to `internal/postmarkclient` and `internal/operatorclient` rather than pulling in `cloudflare-go`. Needs: create Custom Hostname, get Custom Hostname (for the re-check-on-read pattern).
- **New config**: `CLOUDFLARE_API_TOKEN`, `CLOUDFLARE_ZONE_ID`, `CLOUDFLARE_FALLBACK_ORIGIN` (the hostname Cloudflare proxies matching traffic to — see Open decisions).
- **New migration**: a `custom_domains` table, one per instance (`instance_id uuid UNIQUE REFERENCES instances`), shaped like `sender_identities`: `domain`, `cloudflare_hostname_id`, `status` (`pending` | `active` | `failed`), `created_at`. DNS records to publish reuse the existing `dns_records` table with new `record_type` values (`custom_domain_cname`, `custom_domain_ownership`) rather than inventing a parallel records table.
- **New service methods** on `provisioning.Service`, mirroring `AddSenderDomain`/`GetSenderIdentity`/`DeleteSenderIdentity`: `AddCustomDomain`, `GetCustomDomain` (re-checks Cloudflare on read), `DeleteCustomDomain` (removes the Cloudflare hostname, then calls the new operator root-url endpoint to revert).
- **New routes**: `GET`/`POST`/`DELETE /v1/orgs/{orgID}/instances/{instanceID}/custom-domain`.
- **New frontend component**, `CustomDomainCard`, structurally identical to `SenderIdentityCard`: add-domain form → DNS records to publish with copy buttons → status badge → remove action behind a confirmation dialog. Slots into `instance-detail-page.tsx` alongside the existing `SenderIdentityCard`/`PostmarkServerCard`.

## Open decisions

- **Where does Cloudflare's fallback origin actually point?** `docs/plan.md`'s original research assumed a Cloudflare Tunnel in front of a k3s cluster; production today is (per the Dockerfiles/`nginx.conf`) a plain Docker deployment, not k3s. Need to confirm the real current edge setup before picking `CLOUDFLARE_FALLBACK_ORIGIN` — if Cloudflare already proxies straight to the Docker host's public IP/hostname (orange-cloud DNS record, no Tunnel), the fallback origin is just that same hostname and no Tunnel work is needed at all. If it's a Tunnel, the fallback-origin-via-Tunnel pattern from `plan.md` still applies as originally researched.
- **Does `{slug}.{root_domain}` stay reachable once a custom domain is set, or does the custom domain replace it?** Recommend: stays reachable (simpler — `root_url` is a single value listmonk uses for generating links/redirects, but the *subdomain* resolution path in `internal/tenant/resolve.go` doesn't need to be turned off just because a custom domain also resolves). Needs confirming against how the fork actually uses `app.root_url` beyond generating links (e.g. OIDC `redirect_uri`, per `cmd/init.go:1175` — a custom domain changing this needs the org's OIDC app config, if any, updated to match, or it breaks their own SSO).
- **Verification: rely solely on Cloudflare's own domain-control validation, or require our own separate ownership check first?** Recommend Cloudflare's own — it's already a real ownership check (DCV), a second one is redundant complexity.
- **Cost**: Cloudflare for SaaS is 100 custom hostnames free per plan (per `docs/plan.md`'s existing research), $0.10/hostname/month past that — a real per-org cost line only past 100 orgs with a custom domain, not a blocker for v1.1.
- **Polling vs. webhook for verification**: start with re-check-on-read (no new infra, matches `GetSenderIdentity`'s existing pattern); revisit a background poller or Cloudflare webhook only if orgs report the domain not visibly flipping to active promptly enough.

## Build order (fits after `docs/plan.md`'s existing steps 1–5)

1. Confirm the fallback-origin open decision above — blocks picking `CLOUDFLARE_FALLBACK_ORIGIN` and therefore blocks everything else.
2. listmonk fork: the operator root-url endpoint (small, no behavior-path risk) and the tenant-resolution custom-domain lookup (the one piece worth extra review, per above).
3. listnun: `internal/cloudflareclient`, migration, service methods, HTTP routes — live-verified against a real Cloudflare zone the way every other integration in this app has been, not just `go build` clean.
4. Frontend: `CustomDomainCard`.
