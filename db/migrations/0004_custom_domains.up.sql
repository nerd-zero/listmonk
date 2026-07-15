-- Lets an instance be reached at a domain the org owns (e.g. mail.acme.com)
-- instead of only {slug}.{root_domain} -- see docs/custom-domains.md.
-- Exactly one per instance, same shape as sender_identities: `domain` is
-- globally unique so two orgs can never both claim it, and status tracks
-- Cloudflare's own verification (a Custom Hostname) independently of
-- whether the listmonk tenant's app.root_url has actually been flipped to
-- it yet (internal/provisioning.GetCustomDomain only does that once status
-- reaches 'active').
CREATE TABLE custom_domains (
    id                     uuid PRIMARY KEY,
    instance_id            uuid NOT NULL UNIQUE REFERENCES instances (id) ON DELETE CASCADE,
    domain                 text NOT NULL,
    cloudflare_hostname_id text NOT NULL,
    status                 text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'active', 'failed')),
    created_at             timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX custom_domains_domain_key ON custom_domains (lower(domain));

-- dns_records.record_type gains two new values published for a custom
-- domain (the CNAME the org points at Cloudflare's fallback origin, and
-- Cloudflare's ownership-verification TXT) alongside the existing
-- 'dkim' | 'return_path' sender-identity ones -- same table, distinguished
-- by record_type, per docs/custom-domains.md's "Work required" section.
-- No column/constraint change needed: dns_records.record_type has always
-- been a free-form text column, not a CHECK-constrained enum.
