package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

// V6_13_0 makes organizations.name unique - two organizations with the
// same name would otherwise silently coexist, with no way to tell them
// apart in the Operator API's list/lookup endpoints.
func V6_13_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	_, err := db.Exec(`
		ALTER TABLE organizations DROP CONSTRAINT IF EXISTS organizations_name_key;
		ALTER TABLE organizations ADD CONSTRAINT organizations_name_key UNIQUE (name);
	`)
	return err
}
