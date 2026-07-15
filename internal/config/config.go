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

	// PostmarkAccountToken authenticates internal/postmarkclient against
	// Postmark's account-level Account API (create a server + domain per
	// tenant). Optional, like ZitadelServiceAccountKeyPath above: leave
	// blank to run without it -- CreateInstance just skips the
	// provision_postmark_server step, same fail-open pattern as invites.
	PostmarkAccountToken string
	// PostmarkTokenEncryptionKey encrypts postmark_servers.api_token_encrypted
	// at rest (see internal/cryptoutil). A base64-encoded 32-byte AES-256
	// key, e.g. from `openssl rand -base64 32`. Required only if
	// PostmarkAccountToken is set.
	PostmarkTokenEncryptionKey string
	// PostmarkSharedDomainRoot is the parent domain an org with no domain
	// of its own can opt into instead of bringing one: an instance with
	// slug "acme" gets acme.<this>, with the DKIM record published to our
	// own DNS zone -- see internal/provisioning.AddPlatformDomain.
	PostmarkSharedDomainRoot string

	// CloudflareFallbackOrigin is our own hostname, proxied through
	// Cloudflare, that every org CNAMEs their custom domain at -- the
	// org's own DNS never has to be on Cloudflare, since a CNAME works the
	// same regardless of registrar. This is what actually turns the
	// custom-domains feature on: leave it blank to run without it --
	// AddCustomDomain just returns ErrCloudflareNotConfigured. See
	// docs/custom-domains.md.
	//
	// Right now this alone is enough: provisioning.Service only verifies
	// the org's CNAME via plain DNS (Cloudflare's "SSL for SaaS" isn't
	// enabled on our zone yet, so no certificate is issued and a custom
	// domain isn't served over HTTPS yet). CloudflareAPIToken/ZoneID below
	// are a further-optional upgrade, unused until that's wired back in.
	CloudflareFallbackOrigin string
	// CloudflareAPIToken authenticates internal/cloudflareclient against
	// Cloudflare's Custom Hostnames API ("Cloudflare for SaaS"). Currently
	// unused by provisioning.Service (see CloudflareFallbackOrigin above)
	// -- safe to leave blank.
	CloudflareAPIToken string
	// CloudflareZoneID is the zone CloudflareAPIToken above would register
	// each org's domain against, once that path is wired back in.
	CloudflareZoneID string
}

func Load() Config {
	issuer := os.Getenv("ZITADEL_ISSUER")
	domain, insecurePort := parseZitadelHost(issuer)
	return Config{
		DatabaseURL:                  getenv("DATABASE_URL", "postgres://listnun:listnun@localhost:5432/listnun?sslmode=disable"),
		HTTPAddr:                     getenv("HTTP_ADDR", ":8181"),
		ZitadelIssuer:                issuer,
		ZitadelClientID:              os.Getenv("ZITADEL_CLIENT_ID"),
		ZitadelServiceAccountKeyPath: os.Getenv("ZITADEL_SERVICE_ACCOUNT_KEY_PATH"),
		ZitadelOrgID:                 os.Getenv("ZITADEL_ORG_ID"),
		ZitadelDomain:                getenv("ZITADEL_DOMAIN", domain),
		ZitadelInsecurePort:          insecurePort,
		ListmonkOperatorBaseURL:      getenv("LISTMONK_OPERATOR_BASE_URL", "http://localhost:9000"),
		ListmonkOperatorToken:        os.Getenv("LISTMONK_OPERATOR_TOKEN"),
		ListmonkRootDomain:           getenv("LISTMONK_ROOT_DOMAIN", "listmonk.test"),
		PostmarkAccountToken:         os.Getenv("POSTMARK_ACCOUNT_TOKEN"),
		PostmarkTokenEncryptionKey:   os.Getenv("POSTMARK_TOKEN_ENCRYPTION_KEY"),
		PostmarkSharedDomainRoot:     getenv("POSTMARK_SHARED_DOMAIN_ROOT", "mail.listnun.app"),
		CloudflareAPIToken:           os.Getenv("CLOUDFLARE_API_TOKEN"),
		CloudflareZoneID:             os.Getenv("CLOUDFLARE_ZONE_ID"),
		CloudflareFallbackOrigin:     os.Getenv("CLOUDFLARE_FALLBACK_ORIGIN"),
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
