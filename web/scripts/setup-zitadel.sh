#!/bin/bash

# Provisions the "listnun" project + a User Agent (SPA/PKCE) OIDC app, and
# an active SMTP config pointed at mailhog, in the dev Zitadel from
# ../docker-compose.yml.
#
# The automation PAT used to drive the Management/Admin API is read
# straight off the listnun-zitadel-bootstrap volume (see
# ZITADEL_FIRSTINSTANCE_PATPATH / ZITADEL_FIRSTINSTANCE_ORG_MACHINE_* in
# docker-compose.yml) -- no console step required. Pass -t/--token to
# target an instance where that volume isn't reachable (e.g. a
# remote/shared Zitadel).
#
# Re-running this script is safe: it looks up the project/app/SMTP config
# first and reuses them instead of creating duplicates.

DEFAULT_ZITADEL_URL="http://localhost:8081"
DEFAULT_ZITADEL_CONTAINER="listnun-zitadel"
DEFAULT_PROJECT_NAME="listnun"
DEFAULT_APP_NAME="web"
DEFAULT_REDIRECT_URI="http://localhost:5173/auth/callback"
DEFAULT_POST_LOGOUT_URI="http://localhost:5173"
# "mailhog", not "listnun-mailhog" -- Zitadel resolves this from inside the
# compose network, where services reach each other by service name (the
# compose file's key), not container_name. See docker-compose.yml.
DEFAULT_SMTP_HOST="mailhog:1025"
DEFAULT_SMTP_SENDER_ADDRESS="no-reply@listnun.test"
DEFAULT_SMTP_SENDER_NAME="listnun (dev)"

usage() {
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  -t, --token         Zitadel PAT (default: auto-read from the ${DEFAULT_ZITADEL_CONTAINER} container's bootstrap volume)"
    echo "  -z, --zitadel-url   Zitadel base URL (default: ${DEFAULT_ZITADEL_URL})"
    echo "  -c, --container     Docker container to read the automation PAT from (default: ${DEFAULT_ZITADEL_CONTAINER})"
    echo "  -j, --project       Project name (default: ${DEFAULT_PROJECT_NAME})"
    echo "  -a, --app-name      App name (default: ${DEFAULT_APP_NAME})"
    echo "  -r, --redirect-uri  Redirect URI (default: ${DEFAULT_REDIRECT_URI})"
    echo "  -l, --logout-uri    Post-logout redirect URI (default: ${DEFAULT_POST_LOGOUT_URI})"
    echo "  --smtp-host         SMTP host:port (default: ${DEFAULT_SMTP_HOST})"
    echo "  --no-smtp           Skip SMTP setup entirely"
    echo "  -h, --help          Show this help"
    exit 1
}

ZITADEL_URL="$DEFAULT_ZITADEL_URL"
ZITADEL_CONTAINER="$DEFAULT_ZITADEL_CONTAINER"
PROJECT_NAME="$DEFAULT_PROJECT_NAME"
APP_NAME="$DEFAULT_APP_NAME"
REDIRECT_URI="$DEFAULT_REDIRECT_URI"
POST_LOGOUT_URI="$DEFAULT_POST_LOGOUT_URI"
SMTP_HOST="$DEFAULT_SMTP_HOST"
SKIP_SMTP=false
PAT=""

while [[ $# -gt 0 ]]; do
    case "$1" in
        -t|--token) PAT="$2"; shift 2 ;;
        -z|--zitadel-url) ZITADEL_URL="$2"; shift 2 ;;
        -c|--container) ZITADEL_CONTAINER="$2"; shift 2 ;;
        -j|--project) PROJECT_NAME="$2"; shift 2 ;;
        -a|--app-name) APP_NAME="$2"; shift 2 ;;
        -r|--redirect-uri) REDIRECT_URI="$2"; shift 2 ;;
        -l|--logout-uri) POST_LOGOUT_URI="$2"; shift 2 ;;
        --smtp-host) SMTP_HOST="$2"; shift 2 ;;
        --no-smtp) SKIP_SMTP=true; shift ;;
        -h|--help) usage ;;
        *) echo "Unknown option: $1"; usage ;;
    esac
done

check_response() {
    local response="$1"
    local context="$2"
    if echo "$response" | grep -q '"code"'; then
        echo "Error in ${context}: ${response}"
        exit 1
    fi
}

if [[ -z "$PAT" ]]; then
    echo "==> Reading automation PAT from the '${ZITADEL_CONTAINER}' container..."
    PAT_FILE=$(mktemp)
    trap 'rm -f "$PAT_FILE"' EXIT
    if ! docker cp "${ZITADEL_CONTAINER}:/zitadel/bootstrap/automation.pat" "$PAT_FILE" >/dev/null 2>&1; then
        echo "Error: couldn't read /zitadel/bootstrap/automation.pat from container '${ZITADEL_CONTAINER}'."
        echo "Either pass -t/--token explicitly, or make sure the dev stack in ../docker-compose.yml is up"
        echo "and was freshly initialized (the bootstrap PAT is only written on a new instance's first boot)."
        exit 1
    fi
    PAT=$(cat "$PAT_FILE")
fi

AUTH_HEADER="Authorization: Bearer ${PAT}"
MGMT_API="${ZITADEL_URL}/management/v1"

echo "==> Looking up project '${PROJECT_NAME}'..."
PROJECT_SEARCH=$(curl -sS -X POST "${MGMT_API}/projects/_search" \
    -H "$AUTH_HEADER" -H "Content-Type: application/json" \
    -d "{\"queries\":[{\"nameQuery\":{\"name\":\"${PROJECT_NAME}\",\"method\":\"TEXT_QUERY_METHOD_EQUALS\"}}]}")
PROJECT_ID=$(echo "$PROJECT_SEARCH" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

if [[ -z "$PROJECT_ID" ]]; then
    echo "==> Creating project '${PROJECT_NAME}'..."
    PROJECT_RESP=$(curl -sS -X POST "${MGMT_API}/projects" \
        -H "$AUTH_HEADER" -H "Content-Type: application/json" \
        -d "{\"name\": \"${PROJECT_NAME}\"}")
    check_response "$PROJECT_RESP" "create project"
    PROJECT_ID=$(echo "$PROJECT_RESP" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
fi
echo "   Project ID: ${PROJECT_ID}"

echo "==> Looking up application '${APP_NAME}'..."
APP_SEARCH=$(curl -sS -X POST "${MGMT_API}/projects/${PROJECT_ID}/apps/_search" \
    -H "$AUTH_HEADER" -H "Content-Type: application/json" \
    -d "{\"queries\":[{\"nameQuery\":{\"name\":\"${APP_NAME}\",\"method\":\"TEXT_QUERY_METHOD_EQUALS\"}}]}")
APP_ID=$(echo "$APP_SEARCH" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

if [[ -n "$APP_ID" ]]; then
    echo "   Found existing app, id ${APP_ID} -- fetching its client id..."
    APP_DETAIL=$(curl -sS "${MGMT_API}/projects/${PROJECT_ID}/apps/${APP_ID}" -H "$AUTH_HEADER")
    CLIENT_ID=$(echo "$APP_DETAIL" | grep -o '"clientId":"[^"]*"' | head -1 | cut -d'"' -f4)

    # Re-apply the OIDC config on every run so config changes to this script
    # (e.g. the accessTokenType fix -- see git history) land on an
    # already-provisioned app, not just freshly-created ones.
    echo "   Re-applying OIDC config to keep it in sync with this script..."
    UPDATE_RESP=$(curl -sS -X PUT "${MGMT_API}/projects/${PROJECT_ID}/apps/${APP_ID}/oidc_config" \
        -H "$AUTH_HEADER" -H "Content-Type: application/json" \
        -d "{
            \"redirectUris\": [\"${REDIRECT_URI}\"],
            \"responseTypes\": [\"OIDC_RESPONSE_TYPE_CODE\"],
            \"grantTypes\": [\"OIDC_GRANT_TYPE_AUTHORIZATION_CODE\"],
            \"appType\": \"OIDC_APP_TYPE_USER_AGENT\",
            \"authMethodType\": \"OIDC_AUTH_METHOD_TYPE_NONE\",
            \"postLogoutRedirectUris\": [\"${POST_LOGOUT_URI}\"],
            \"devMode\": true,
            \"accessTokenType\": \"OIDC_TOKEN_TYPE_JWT\",
            \"accessTokenRoleAssertion\": false,
            \"idTokenRoleAssertion\": false,
            \"idTokenUserinfoAssertion\": true
        }")
    check_response "$UPDATE_RESP" "update application oidc config"
else
    echo "==> Creating User Agent application '${APP_NAME}'..."
    APP_RESP=$(curl -sS -X POST "${MGMT_API}/projects/${PROJECT_ID}/apps/oidc" \
        -H "$AUTH_HEADER" -H "Content-Type: application/json" \
        -d "{
            \"name\": \"${APP_NAME}\",
            \"redirectUris\": [\"${REDIRECT_URI}\"],
            \"responseTypes\": [\"OIDC_RESPONSE_TYPE_CODE\"],
            \"grantTypes\": [\"OIDC_GRANT_TYPE_AUTHORIZATION_CODE\"],
            \"appType\": \"OIDC_APP_TYPE_USER_AGENT\",
            \"authMethodType\": \"OIDC_AUTH_METHOD_TYPE_NONE\",
            \"postLogoutRedirectUris\": [\"${POST_LOGOUT_URI}\"],
            \"version\": \"OIDC_VERSION_1_0\",
            \"devMode\": true,
            \"accessTokenType\": \"OIDC_TOKEN_TYPE_JWT\",
            \"accessTokenRoleAssertion\": false,
            \"idTokenRoleAssertion\": false,
            \"idTokenUserinfoAssertion\": true,
            \"additionalOrigins\": []
        }")
    check_response "$APP_RESP" "create application"
    CLIENT_ID=$(echo "$APP_RESP" | grep -o '"clientId":"[^"]*"' | cut -d'"' -f4)
fi

if [[ -z "$CLIENT_ID" ]]; then
    echo "   Couldn't determine the client id -- last response:"
    echo "${APP_DETAIL:-$APP_RESP}"
    exit 1
fi
echo "   Client ID: ${CLIENT_ID}"

if [[ "$SKIP_SMTP" != "true" ]]; then
    ADMIN_API="${ZITADEL_URL}/admin/v1"

    echo "==> Checking SMTP configuration..."
    SMTP_SEARCH=$(curl -sS -X POST "${ADMIN_API}/smtp/_search" \
        -H "$AUTH_HEADER" -H "Content-Type: application/json" -d '{}')
    SMTP_ID=$(echo "$SMTP_SEARCH" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    SMTP_STATE=$(echo "$SMTP_SEARCH" | grep -o '"state":"[^"]*"' | head -1 | cut -d'"' -f4)

    if [[ -z "$SMTP_ID" ]]; then
        echo "==> Creating SMTP config pointed at ${SMTP_HOST}..."
        SMTP_RESP=$(curl -sS -X POST "${ADMIN_API}/smtp" \
            -H "$AUTH_HEADER" -H "Content-Type: application/json" \
            -d "{
                \"senderAddress\": \"${DEFAULT_SMTP_SENDER_ADDRESS}\",
                \"senderName\": \"${DEFAULT_SMTP_SENDER_NAME}\",
                \"tls\": false,
                \"host\": \"${SMTP_HOST}\",
                \"user\": \"\",
                \"password\": \"\"
            }")
        check_response "$SMTP_RESP" "create SMTP config"
        SMTP_ID=$(echo "$SMTP_RESP" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
        SMTP_STATE=""
    fi

    if [[ "$SMTP_STATE" != "SMTP_CONFIG_ACTIVE" ]]; then
        echo "==> Activating SMTP config ${SMTP_ID}..."
        curl -sS -X POST "${ADMIN_API}/smtp/${SMTP_ID}/_activate" -H "$AUTH_HEADER" >/dev/null
    fi
    echo "   SMTP: ${SMTP_HOST} -- catches everything at http://localhost:8025 (mailhog)"
fi

echo ""
echo "==> Done. Set the following in web/.env.local:"
echo ""
echo "   VITE_OIDC_AUTHORITY=${ZITADEL_URL}"
echo "   VITE_OIDC_CLIENT_ID=${CLIENT_ID}"
echo "   VITE_OIDC_REDIRECT_URI=${REDIRECT_URI}"
echo "   VITE_OIDC_POST_LOGOUT_REDIRECT_URI=${POST_LOGOUT_URI}"
