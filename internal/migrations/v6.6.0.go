package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

// V6_6_0 re-scopes the `settings` table's uniqueness from a single global
// `UNIQUE(key)` to a composite `PRIMARY KEY (tenant_id, key)`, part of
// phase 5 (issue #32): settings become fully per-tenant, no global/
// per-tenant split. This was deliberately deferred from phase 1
// (v6.4.0) - re-scoping it there, before queries/misc.sql's
// get-settings/update-settings/update-settings-by-key queries were
// updated to filter by tenant_id in the same change, would have made
// every settings read/write silently operate across all tenants at once
// (the old `WHERE s.key = c.key` update-settings query has no tenant
// filter yet) rather than erroring loudly - worse than the ON CONFLICT
// breakage phase 1 was originally worried about.
func V6_6_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	_, err := db.Exec(`
		ALTER TABLE settings DROP CONSTRAINT IF EXISTS settings_key_key;

		DO $$ BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conrelid = 'settings'::regclass AND contype = 'p'
			) THEN
				ALTER TABLE settings ADD PRIMARY KEY (tenant_id, key);
			END IF;
		END $$;
	`)
	return err
}
