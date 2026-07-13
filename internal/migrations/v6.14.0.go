package migrations

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
	"github.com/lib/pq"
)

// operatorGrantTables lists every table the Operator API's BYPASSRLS
// connection needs to read across tenants: the two cross-tenant management
// tables plus every RLS-scoped table in schema.sql. BYPASSRLS only skips
// row-level security policies - it does not grant table-level access, so a
// non-owner role still needs an explicit GRANT before cross-tenant reads
// will work.
var operatorGrantTables = []string{
	"tenants", "organizations",
	"subscribers", "lists", "templates", "campaigns", "media", "links",
	"bounces", "roles", "users", "subscriber_lists", "campaign_lists",
	"campaign_views", "campaign_media", "link_clicks", "settings",
}

// ApplyOperatorGrants grants the configured [operator] BYPASSRLS role read
// access to every tenant table, and arranges for it to keep that access on
// tables created by future migrations. It's a no-op when operator.db_user
// is unset - the default, single-tenant case - since this runs on every
// deployment of this fork, not just ones using the operator feature. It
// also skips (rather than fails) if the role hasn't been provisioned in
// Postgres yet, since that's managed out-of-band by infra and may not
// exist on the app's first boot; the caller is expected to retry (both
// call sites below do, on every --upgrade / --install run).
//
// Exported and called from two places: V6_14_0 below (deployments
// upgrading from an older release) and cmd/install.go's fresh-install path
// directly. A fresh --install marks the migration ledger as already caught
// up to migList's newest entry - which, once this file is registered, is
// V6_14_0 itself - so it never actually runs via the upgrade path on a new
// database. schema.sql can't express this GRANT either, since the role
// name is only known at runtime via config. Calling it unconditionally
// after install closes that gap.
func ApplyOperatorGrants(db *sqlx.DB, ko *koanf.Koanf, lo *log.Logger) error {
	role := ko.String("operator.db_user")
	if role == "" {
		return nil
	}

	var exists bool
	if err := db.Get(&exists, `SELECT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = $1)`, role); err != nil {
		return err
	}
	if !exists {
		lo.Printf("operator.db_user %q is configured but the role does not exist in Postgres yet; skipping operator GRANTs (will retry on next upgrade/install)", role)
		return nil
	}

	quotedRole := pq.QuoteIdentifier(role)
	for _, t := range operatorGrantTables {
		if _, err := db.Exec(fmt.Sprintf(`GRANT SELECT ON %s TO %s`, pq.QuoteIdentifier(t), quotedRole)); err != nil {
			return err
		}
	}

	if _, err := db.Exec(fmt.Sprintf(`ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO %s`, quotedRole)); err != nil {
		return err
	}

	return nil
}

// V6_14_0 applies the operator role's GRANTs for deployments upgrading from
// an older release. See ApplyOperatorGrants for why fresh installs need a
// separate call site instead of relying on this migration running.
func V6_14_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	return ApplyOperatorGrants(db, ko, lo)
}
