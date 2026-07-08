package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

// V6_7_0 adds a tenant_id dimension to the three materialized views that
// back dashboard counts/charts and per-list subscriber stats
// (mat_dashboard_counts, mat_dashboard_charts, mat_list_subscriber_stats).
//
// Found while closing out issue #40 (internal/core tenantID-threading):
// these views each computed a single global row (or, for
// mat_list_subscriber_stats, a single global list_id=0 "all subscribers"
// row) with no tenant_id column at all. queries/subscribers.sql's
// query-subscribers-count-all falls back to that global list_id=0 row
// whenever a request has no list filter - the common case - so every
// tenant's unfiltered "total subscribers" count was silently the sum
// across ALL tenants, not just their own. Same issue for the dashboard
// totals/charts. This is a real cross-tenant data leak, not just the
// generic "refresh cost" performance question already tracked in
// docs/design/multi-tenancy.md's Open Questions.
//
// Each view now has one row per tenant (driven by a join/group against
// the tenants table), with its unique index widened to lead with
// tenant_id (required for REFRESH MATERIALIZED VIEW CONCURRENTLY, which
// internal/core.Core.RefreshMatView already uses). The refresh mechanism
// itself is unchanged - REFRESH MATERIALIZED VIEW refreshes the whole
// view (all tenants) in one statement, matching the Open Questions
// section's recommended default of keeping the existing global refresh
// cadence and filtering by tenant_id at query time.
func V6_7_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	_, err := db.Exec(`
		DROP MATERIALIZED VIEW IF EXISTS mat_dashboard_counts;
		CREATE MATERIALIZED VIEW mat_dashboard_counts AS
			SELECT NOW() AS updated_at, t.id AS tenant_id,
				JSON_BUILD_OBJECT(
					'subscribers', JSON_BUILD_OBJECT(
						'total', (SELECT COUNT(*) FROM subscribers WHERE tenant_id = t.id),
						'blocklisted', (SELECT COUNT(*) FROM subscribers WHERE tenant_id = t.id AND status = 'blocklisted'),
						'orphans', (
							SELECT COUNT(subscribers.id) FROM subscribers
							LEFT JOIN subscriber_lists ON (subscribers.id = subscriber_lists.subscriber_id)
							WHERE subscribers.tenant_id = t.id AND subscriber_lists.subscriber_id IS NULL
						)
					),
					'lists', JSON_BUILD_OBJECT(
						'total', (SELECT COUNT(*) FROM lists WHERE tenant_id = t.id),
						'private', (SELECT COUNT(*) FROM lists WHERE tenant_id = t.id AND type = 'private'),
						'public', (SELECT COUNT(*) FROM lists WHERE tenant_id = t.id AND type = 'public'),
						'optin_single', (SELECT COUNT(*) FROM lists WHERE tenant_id = t.id AND optin = 'single'),
						'optin_double', (SELECT COUNT(*) FROM lists WHERE tenant_id = t.id AND optin = 'double')
					),
					'campaigns', JSON_BUILD_OBJECT(
						'total', (SELECT COUNT(*) FROM campaigns WHERE tenant_id = t.id),
						'by_status', (
							SELECT COALESCE(JSON_OBJECT_AGG(status, num), '{}'::JSON) FROM
							(SELECT status, COUNT(*) AS num FROM campaigns WHERE tenant_id = t.id GROUP BY status) r
						)
					),
					'messages', (SELECT COALESCE(SUM(sent), 0) FROM campaigns WHERE tenant_id = t.id)
				) AS data
			FROM tenants t;
		DROP INDEX IF EXISTS mat_dashboard_stats_idx;
		CREATE UNIQUE INDEX mat_dashboard_stats_idx ON mat_dashboard_counts (tenant_id);

		DROP MATERIALIZED VIEW IF EXISTS mat_dashboard_charts;
		CREATE MATERIALIZED VIEW mat_dashboard_charts AS
			SELECT NOW() AS updated_at, t.id AS tenant_id,
				JSON_BUILD_OBJECT(
					'link_clicks', COALESCE((
						SELECT JSON_AGG(ROW_TO_JSON(row)) FROM (
							WITH viewDates AS (
								SELECT created_at::DATE AS to_date,
									   created_at::DATE - INTERVAL '30 DAY' AS from_date
								FROM link_clicks WHERE tenant_id = t.id ORDER BY id DESC LIMIT 1
							)
							SELECT COUNT(*) AS count, created_at::DATE AS date FROM link_clicks
							WHERE tenant_id = t.id
								AND created_at >= (SELECT from_date FROM viewDates)
								AND created_at < (SELECT to_date FROM viewDates) + INTERVAL '1 day'
							GROUP BY date ORDER BY date
						) row
					), '[]'),
					'campaign_views', COALESCE((
						SELECT JSON_AGG(ROW_TO_JSON(row)) FROM (
							WITH viewDates AS (
								SELECT created_at::DATE AS to_date,
									   created_at::DATE - INTERVAL '30 DAY' AS from_date
								FROM campaign_views WHERE tenant_id = t.id ORDER BY id DESC LIMIT 1
							)
							SELECT COUNT(*) AS count, created_at::DATE AS date FROM campaign_views
							WHERE tenant_id = t.id
								AND created_at >= (SELECT from_date FROM viewDates)
								AND created_at < (SELECT to_date FROM viewDates) + INTERVAL '1 day'
							GROUP BY date ORDER BY date
						) row
					), '[]')
				) AS data
			FROM tenants t;
		DROP INDEX IF EXISTS mat_dashboard_charts_idx;
		CREATE UNIQUE INDEX mat_dashboard_charts_idx ON mat_dashboard_charts (tenant_id);

		DROP MATERIALIZED VIEW IF EXISTS mat_list_subscriber_stats;
		CREATE MATERIALIZED VIEW mat_list_subscriber_stats AS
			SELECT NOW() AS updated_at, lists.tenant_id AS tenant_id, lists.id AS list_id,
				subscriber_lists.status, COUNT(subscriber_lists.status) AS subscriber_count FROM lists
			LEFT JOIN subscriber_lists ON (subscriber_lists.list_id = lists.id)
			GROUP BY lists.tenant_id, lists.id, subscriber_lists.status
			UNION ALL
			SELECT NOW() AS updated_at, subscribers.tenant_id AS tenant_id, 0 AS list_id,
				NULL AS status, COUNT(id) AS subscriber_count FROM subscribers
			GROUP BY subscribers.tenant_id;
		DROP INDEX IF EXISTS mat_list_subscriber_stats_idx;
		CREATE UNIQUE INDEX mat_list_subscriber_stats_idx ON mat_list_subscriber_stats (tenant_id, list_id, status);
	`)
	return err
}
