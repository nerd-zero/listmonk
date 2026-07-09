#!/bin/bash

# Provisions the "listnun" project + a User Agent (SPA/PKCE) OIDC app in the
# dev Zitadel from ../docker-compose.yml, for react-oidc-context to use.
#
# The automation PAT used to drive the Management API is read straight off
# the listnun-zitadel-bootstrap volume (see ZITADEL_FIRSTINSTANCE_PATPATH /
# ZITADEL_FIRSTINSTANCE_ORG_MACHINE_* in docker-compose.yml) -- no console
# step required. Pass -t/--token to target an instance where that volume
# isn't reachable (e.g. a remote/shared Zitadel).
#
# Re-running this script is safe: it looks up the project/app by name first
# and reuses them instead of creating duplicates.

DEFAULT_ZITADEL_URL="http://localhost:8081"
DEFAULT_ZITADEL_CONTAINER="listnun-zitadel"
DEFAULT_PROJECT_NAME="listnun"
DEFAULT_APP_NAME="web"
DEFAULT_REDIRECT_URI="http://localhost:5173/auth/callback"
DEFAULT_POST_LOGOUT_URI="http://localhost:5173"

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
    echo "  -h, --help          Show this help"
    exit 1
}

ZITADEL_URL="$DEFAULT_ZITADEL_URL"
ZITADEL_CONTAINER="$DEFAULT_ZITADEL_CONTAINER"
PROJECT_NAME="$DEFAULT_PROJECT_NAME"
APP_NAME="$DEFAULT_APP_NAME"
REDIRECT_URI="$DEFAULT_REDIRECT_URI"
POST_LOGOUT_URI="$DEFAULT_POST_LOGOUT_URI"
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
            \"accessTokenType\": \"OIDC_TOKEN_TYPE_BEARER\",
            \"accessTokenRoleAssertion\": false,
            \"idTokenRoleAssertion\": false,
            \"idTokenUserinfoAssertion\": false,
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

echo ""
echo "==> Done. Set the following in web/.env.local:"
echo ""
echo "   VITE_OIDC_AUTHORITY=${ZITADEL_URL}"
echo "   VITE_OIDC_CLIENT_ID=${CLIENT_ID}"
echo "   VITE_OIDC_REDIRECT_URI=${REDIRECT_URI}"
echo "   VITE_OIDC_POST_LOGOUT_REDIRECT_URI=${POST_LOGOUT_URI}"
