ALTER TABLE sender_identities DROP CONSTRAINT sender_identities_kind_check;
ALTER TABLE sender_identities ADD CONSTRAINT sender_identities_kind_check
    CHECK (kind IN ('domain', 'sender_signature'));
