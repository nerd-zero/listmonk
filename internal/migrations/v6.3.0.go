package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

func V6_3_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	// Add `integration_id` to the `scrub` settings JSON if missing.
	// Also upsert the row with all required fields in case it doesn't exist yet.
	if _, err := db.Exec(`
		INSERT INTO settings (key, value)
		VALUES ('scrub', '{"enabled": false, "url": "", "api_key": "", "integration_id": 0}')
		ON CONFLICT (key) DO UPDATE
		SET value = JSONB_SET(settings.value, '{integration_id}',
			COALESCE(settings.value->'integration_id', '0'::JSONB))
		WHERE NOT (settings.value ? 'integration_id');
	`); err != nil {
		return err
	}

	return nil
}
