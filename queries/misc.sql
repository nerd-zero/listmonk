-- name: get-dashboard-charts
SELECT data FROM mat_dashboard_charts WHERE tenant_id = $1;

-- name: get-dashboard-counts
SELECT data FROM mat_dashboard_counts WHERE tenant_id = $1;

-- name: get-settings
SELECT JSON_OBJECT_AGG(key, value) AS settings FROM (SELECT * FROM settings WHERE tenant_id = $1 ORDER BY key) t;

-- name: update-settings
UPDATE settings AS s SET value = c.value
    -- For each key in the incoming JSON map, update the row with the key and its value.
    -- The tenant_id filter is essential here, not just defense-in-depth: key names repeat
    -- across tenants by design, so without it this would update every tenant's row that
    -- happens to share a key name in the incoming map.
    FROM(SELECT * FROM JSONB_EACH($1)) AS c(key, value) WHERE s.key = c.key AND s.tenant_id = $2;

-- name: update-settings-by-key
UPDATE settings SET value = $2, updated_at = NOW() WHERE key = $1 AND tenant_id = $3;

-- name: get-db-info
SELECT JSON_BUILD_OBJECT('version', (SELECT VERSION()),
                        'size_mb', (SELECT ROUND(pg_database_size((SELECT CURRENT_DATABASE()))/(1024^2)))) AS info;
