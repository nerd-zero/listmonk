DROP TABLE IF EXISTS sender_identities;
ALTER TABLE postmark_servers ADD COLUMN sending_domain text NOT NULL DEFAULT '';
ALTER TABLE postmark_servers ALTER COLUMN sending_domain DROP DEFAULT;
