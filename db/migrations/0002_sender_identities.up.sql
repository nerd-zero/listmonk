-- sending_domain used to be auto-derived from the instance slug at
-- provision_postmark_server time (see internal/provisioning). It's now
-- something the org types in themselves via sender_identities below, so
-- postmark_servers goes back to tracking only the Postmark server
-- resource itself.
ALTER TABLE postmark_servers DROP COLUMN sending_domain;

-- A sender identity is how an instance is actually allowed to send mail
-- from Postmark's perspective: either a full sending domain (DKIM +
-- Return-Path, published as dns_records below) or a single verified
-- sender signature (one From address, confirmed by clicking a link
-- Postmark emails directly -- no DNS involved). Exactly one per instance
-- for now; `value` is globally unique so two different orgs can never
-- both claim the same domain or the same sender email.
CREATE TABLE sender_identities (
    id          uuid PRIMARY KEY,
    instance_id uuid NOT NULL UNIQUE REFERENCES instances (id) ON DELETE CASCADE,
    kind        text NOT NULL CHECK (kind IN ('domain', 'sender_signature')),
    value       text NOT NULL, -- the domain name, or the sender email address
    postmark_id text NOT NULL, -- Postmark's Domain ID or SenderSignature ID
    status      text NOT NULL DEFAULT 'pending', -- 'pending' | 'confirmed'
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX sender_identities_value_key ON sender_identities (lower(value));
