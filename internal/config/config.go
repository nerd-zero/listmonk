package config

import (
	"net"
	"os"
	"strconv"
	"strings"
)

// Config holds environment-provided settings shared by cmd/api and cmd/worker.
type Config struct {
	DatabaseURL     string
	HTTPAddr        string
	ZitadelIssuer   string // e.g. https://<instance>.zitadel.cloud
	ZitadelClientID string

	// ZitadelServiceAccountKeyPath/OrgID authenticate internal/zitadelmgmt as
	// a machine user ("service account") to invite new people directly --
	// distinct from ZitadelIssuer/ClientID above, which only ever verify a
	// human's own bearer token, never act as an identity of their own.
	// Empty by default: inviting users is optional, unlike the operator
	// token below which the API can't run without.
	ZitadelServiceAccountKeyPath string
	ZitadelOrgID                 string
	// ZitadelDomain is the bare host, no port, no scheme (e.g.
	// "your-instance.zitadel.cloud" or "localhost") -- derived from
	// ZitadelIssuer if unset.
	ZitadelDomain string
	// ZitadelInsecurePort is nonzero only for a local plaintext Zitadel
	// (docker-compose.yml's dev stack runs with --tlsMode disabled) --
	// derived from an http:// issuer's port. Zero means use TLS.
	ZitadelInsecurePort uint16

	// ListmonkOperatorBaseURL/Token talk to the listmonk fork's cross-tenant
	// Operator API (internal/operatorclient) -- see
	// docs/design/multi-tenancy.md in that repo. Token has no default: it
	// must come from the same [operator] config as the listmonk fork itself.
	ListmonkOperatorBaseURL string
	ListmonkOperatorToken   string
	// ListmonkRootDomain mirrors the fork's own root_domain -- used only to
	// show a tenant's workspace URL before its own setup_url is known.
	ListmonkRootDomain string
}

func Load() Config {
	issuer := os.Getenv("ZITADEL_ISSUER")
	domain, insecurePort := parseZitadelHost(issuer)
	return Config{
		DatabaseURL:                  getenv("DATABASE_URL", "postgres://listnun:listnun@localhost:5432/listnun?sslmode=disable"),
		HTTPAddr:                     getenv("HTTP_ADDR", ":8080"),
		ZitadelIssuer:                issuer,
		ZitadelClientID:              os.Getenv("ZITADEL_CLIENT_ID"),
		ZitadelServiceAccountKeyPath: os.Getenv("ZITADEL_SERVICE_ACCOUNT_KEY_PATH"),
		ZitadelOrgID:                 os.Getenv("ZITADEL_ORG_ID"),
		ZitadelDomain:                getenv("ZITADEL_DOMAIN", domain),
		ZitadelInsecurePort:          insecurePort,
		ListmonkOperatorBaseURL:      getenv("LISTMONK_OPERATOR_BASE_URL", "http://localhost:9000"),
		ListmonkOperatorToken:        os.Getenv("LISTMONK_OPERATOR_TOKEN"),
		ListmonkRootDomain:           getenv("LISTMONK_ROOT_DOMAIN", "listmonk.test"),
	}
}

// parseZitadelHost splits an issuer URL into the bare host the SDK wants
// and, for a plain-http issuer (local dev only -- see
// docker-compose.yml's --tlsMode disabled Zitadel), the port to connect
// insecurely on. A https issuer always returns insecurePort 0 (use TLS).
func parseZitadelHost(issuer string) (domain string, insecurePort uint16) {
	isHTTP := strings.HasPrefix(issuer, "http://")
	d := strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(issuer, "https://"), "http://"), "/")

	host, port, err := net.SplitHostPort(d)
	if err != nil {
		return d, 0
	}
	if isHTTP {
		if p, err := strconv.Atoi(port); err == nil {
			return host, uint16(p)
		}
	}
	return host, 0
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
