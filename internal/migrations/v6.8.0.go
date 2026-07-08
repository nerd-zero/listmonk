package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

// V6_8_0 widens two pre-existing single-tenant UNIQUE indexes to include
// tenant_id, found while auditing the multi-tenant onboarding flow
// (issue #40's Phase 8 audit):
//
//   - roles(type, name) WHERE name IS NOT NULL: every tenant's first-time
//     setup (cmd/auth.go's doFirstTimeSetup) creates a role literally named
//     "Super Admin". Under the old global index, only the very first
//     tenant ever set up on an installation could create that role -
//     every subsequent tenant's setup hard-failed with a unique
//     constraint violation, completely blocking onboarding past the
//     first tenant.
//   - templates(is_default) WHERE is_default = true: only the first
//     tenant to ever mark a template as default could have one - every
//     other tenant's attempt to set their own default template
//     (queries/templates.sql's set-default-template) also hard-failed,
//     since the constraint had no tenant dimension at all.
//
// Both were additive, single-tenant-era constraints from v4.0.0/early
// schema.sql, predating this fork's tenant_id columns (phase 1, v6.4.0)
// - phase 1 deliberately deferred touching pre-existing constraints (see
// that migration's comments) to avoid breaking ON CONFLICT clauses.
// This is that deferred cleanup, scoped to the two constraints that
// actively block (not just weaken isolation for) legitimate multi-tenant
// usage. Other pre-existing global-uniqueness gaps (subscribers.email,
// links.url, users.username/email) are known, already documented
// elsewhere (see docs/design/multi-tenancy.md), and deliberately left
// out of this migration - they cause soft cross-tenant collisions on
// human-chosen values, not hard onboarding failures, and widening them
// changes ON CONFLICT semantics relied on by existing queries in ways
// that need their own dedicated review.
func V6_8_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	_, err := db.Exec(`
		DROP INDEX IF EXISTS idx_roles_name;
		CREATE UNIQUE INDEX idx_roles_name ON roles (tenant_id, type, name) WHERE name IS NOT NULL;

		DROP INDEX IF EXISTS templates_is_default_idx;
		CREATE UNIQUE INDEX templates_is_default_idx ON templates (tenant_id, is_default) WHERE is_default = true;
	`)
	return err
}
