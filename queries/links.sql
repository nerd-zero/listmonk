-- links
-- name: create-link
-- NOTE: url is globally unique (pre-multi-tenancy constraint, not yet
-- revisited - see docs/design/multi-tenancy.md) so ON CONFLICT reuses
-- whichever tenant registered a given URL first; the tenant_id passed here
-- only applies on first insert.
INSERT INTO links (uuid, url, tenant_id) VALUES($1, $2, $3) ON CONFLICT (url) DO UPDATE SET url=EXCLUDED.url RETURNING uuid;

-- name: get-link-url
SELECT url FROM links WHERE uuid = $1;

-- name: register-link-click
WITH link AS(
    SELECT id, url FROM links WHERE uuid = $1
)
INSERT INTO link_clicks (campaign_id, subscriber_id, link_id, tenant_id) VALUES(
    (SELECT id FROM campaigns WHERE uuid = $2),
    (SELECT id FROM subscribers WHERE
        (CASE WHEN $3::TEXT != '' THEN subscribers.uuid = $3::UUID ELSE FALSE END)
    ),
    (SELECT id FROM link),
    $4
) RETURNING (SELECT url FROM link);
