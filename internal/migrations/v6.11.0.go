package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

// V6_11_0 widens subscribers.email from a global UNIQUE constraint to a
// per-tenant one - the "deferred, has ON CONFLICT dependents" gap flagged
// but deliberately left alone back in phase 1/v6.9.0. Found live via real
// data corruption: queries/subscribers.sql's upsert-subscriber (the bulk
// CSV importer's insert path) does `ON CONFLICT (email) DO UPDATE`
// against the global constraint, so importing a CSV into one tenant that
// happens to contain an e-mail already used by a *different* tenant
// silently attached that other tenant's existing subscriber row to the
// importing tenant's list - subscriber_lists ended up with tenant_id set
// to the importing tenant but subscriber_id pointing at a subscriber
// actually owned by a different tenant. Reproduced by two tenants
// importing overlapping e-mail lists; confirmed via a join between
// subscriber_lists and subscribers on tenant_id mismatch.
//
// Constraint/index names are kept identical (only their column lists
// change) so the existing pq.Error.Constraint == "subscribers_email_key"
// check in internal/core/subscribers.go's CreateSubscriber keeps working
// unchanged - now correctly scoped to "exists within this tenant" rather
// than globally.
func V6_11_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	_, err := db.Exec(`
		ALTER TABLE subscribers DROP CONSTRAINT IF EXISTS subscribers_email_key;
		ALTER TABLE subscribers ADD CONSTRAINT subscribers_email_key UNIQUE (tenant_id, email);

		DROP INDEX IF EXISTS idx_subs_email;
		CREATE UNIQUE INDEX idx_subs_email ON subscribers (tenant_id, LOWER(email));
	`)
	return err
}
