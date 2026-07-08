package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

// V6_9_0 widens users.username/users.email from global UNIQUE constraints
// to per-tenant ones, found live while onboarding a second real tenant
// through the Operator API's setup flow (issue #40's Phase 8 audit had
// already flagged this as a known, deferred gap - unlike the roles/
// templates constraints fixed in v6.8.0, this one didn't block
// onboarding outright, just collided on human-chosen values, so it was
// deliberately left for later review). It turned out to block real
// usage in practice: every tenant's natural choice of "admin" as their
// first user's username collided with whichever tenant claimed it
// first.
//
// Confirmed before writing this migration that no queries/*.sql
// ON CONFLICT clause and no Go error-handling code references
// users_username_key/users_email_key by name (grepped both) - unlike
// subscribers.email or links.url, which do have ON CONFLICT dependents
// and still need their own dedicated review before widening.
func V6_9_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	_, err := db.Exec(`
		ALTER TABLE users DROP CONSTRAINT IF EXISTS users_username_key;
		CREATE UNIQUE INDEX IF NOT EXISTS users_username_key ON users (tenant_id, username);

		ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;
		CREATE UNIQUE INDEX IF NOT EXISTS users_email_key ON users (tenant_id, email);
	`)
	return err
}
