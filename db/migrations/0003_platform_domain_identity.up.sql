-- Adds the third sender identity option: an org with no domain of its own
-- can opt into a subdomain of ours instead (e.g. acme.mail.listnun.app),
-- with the DKIM record published to our own DNS zone rather than the
-- org's -- see internal/provisioning.AddPlatformDomain.
ALTER TABLE sender_identities DROP CONSTRAINT sender_identities_kind_check;
ALTER TABLE sender_identities ADD CONSTRAINT sender_identities_kind_check
    CHECK (kind IN ('domain', 'sender_signature', 'platform_domain'));
