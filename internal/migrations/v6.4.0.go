package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

// V6_4_0 adds multi-tenancy scaffolding: a `tenants` table and a `tenant_id`
// column (defaulting to the seeded tenant 1) on every tenant-scoped table.
// Purely additive - no existing constraint, index, or query behavior
// changes, so this is safe to run ahead of the application code that will
// start threading tenant IDs explicitly in a later phase. Uniqueness
// constraints (subscribers.email, users.username/email, links.url, etc.)
// are intentionally left untouched here: several are targeted by
// `ON CONFLICT` clauses in queries/*.sql (subscribers.sql's
// `ON CONFLICT (email)`, links.sql's `ON CONFLICT (url)`, roles.sql's
// `ON CONFLICT (parent_id, list_id)`) that must be updated in the same
// change that re-scopes their backing constraint - that happens in the
// auth/request-flow phase, not here.
func V6_4_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	_, err := db.Exec(`
		DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'tenant_status') THEN
				CREATE TYPE tenant_status AS ENUM ('active', 'suspended', 'disabled');
			END IF;
		END $$;

		CREATE TABLE IF NOT EXISTS tenants (
			id          SERIAL PRIMARY KEY,
			slug        TEXT NOT NULL UNIQUE,
			name        TEXT NOT NULL,
			status      tenant_status NOT NULL DEFAULT 'active',
			created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		-- Every existing install becomes tenant id=1, "default".
		INSERT INTO tenants (id, slug, name, status)
		VALUES (1, 'default', 'Default', 'active')
		ON CONFLICT (id) DO NOTHING;
		SELECT setval('tenants_id_seq', GREATEST((SELECT MAX(id) FROM tenants), 1));

		ALTER TABLE subscribers      ADD COLUMN IF NOT EXISTS tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE ON UPDATE CASCADE;
		ALTER TABLE lists            ADD COLUMN IF NOT EXISTS tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE ON UPDATE CASCADE;
		ALTER TABLE templates        ADD COLUMN IF NOT EXISTS tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE ON UPDATE CASCADE;
		ALTER TABLE campaigns        ADD COLUMN IF NOT EXISTS tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE ON UPDATE CASCADE;
		ALTER TABLE media            ADD COLUMN IF NOT EXISTS tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE ON UPDATE CASCADE;
		ALTER TABLE links            ADD COLUMN IF NOT EXISTS tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE ON UPDATE CASCADE;
		ALTER TABLE bounces          ADD COLUMN IF NOT EXISTS tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE ON UPDATE CASCADE;
		ALTER TABLE roles            ADD COLUMN IF NOT EXISTS tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE ON UPDATE CASCADE;
		ALTER TABLE users            ADD COLUMN IF NOT EXISTS tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE ON UPDATE CASCADE;
		ALTER TABLE subscriber_lists ADD COLUMN IF NOT EXISTS tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE ON UPDATE CASCADE;
		ALTER TABLE campaign_lists   ADD COLUMN IF NOT EXISTS tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE ON UPDATE CASCADE;
		ALTER TABLE campaign_views   ADD COLUMN IF NOT EXISTS tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE ON UPDATE CASCADE;
		ALTER TABLE campaign_media   ADD COLUMN IF NOT EXISTS tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE ON UPDATE CASCADE;
		ALTER TABLE link_clicks      ADD COLUMN IF NOT EXISTS tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE ON UPDATE CASCADE;
		ALTER TABLE settings         ADD COLUMN IF NOT EXISTS tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE ON UPDATE CASCADE;

		CREATE INDEX IF NOT EXISTS idx_subscribers_tenant      ON subscribers(tenant_id);
		CREATE INDEX IF NOT EXISTS idx_lists_tenant            ON lists(tenant_id);
		CREATE INDEX IF NOT EXISTS idx_templates_tenant        ON templates(tenant_id);
		CREATE INDEX IF NOT EXISTS idx_campaigns_tenant        ON campaigns(tenant_id);
		CREATE INDEX IF NOT EXISTS idx_media_tenant            ON media(tenant_id);
		CREATE INDEX IF NOT EXISTS idx_links_tenant            ON links(tenant_id);
		CREATE INDEX IF NOT EXISTS idx_bounces_tenant          ON bounces(tenant_id);
		CREATE INDEX IF NOT EXISTS idx_roles_tenant            ON roles(tenant_id);
		CREATE INDEX IF NOT EXISTS idx_users_tenant            ON users(tenant_id);
		CREATE INDEX IF NOT EXISTS idx_subscriber_lists_tenant ON subscriber_lists(tenant_id);
		CREATE INDEX IF NOT EXISTS idx_campaign_lists_tenant   ON campaign_lists(tenant_id);
		CREATE INDEX IF NOT EXISTS idx_campaign_views_tenant   ON campaign_views(tenant_id);
		CREATE INDEX IF NOT EXISTS idx_campaign_media_tenant   ON campaign_media(tenant_id);
		CREATE INDEX IF NOT EXISTS idx_link_clicks_tenant      ON link_clicks(tenant_id);
		CREATE INDEX IF NOT EXISTS idx_settings_tenant         ON settings(tenant_id);
	`)
	return err
}
