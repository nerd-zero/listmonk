// package tenant resolves the tenant a request belongs to from its
// subdomain, ahead of authentication, so that both authenticated and
// public routes (unsubscribe, archive, tracking pixel) can be scoped to a
// tenant.
package tenant

import (
	"net"
	"net/http"
	"strings"

	"github.com/knadh/listmonk/internal/core"
	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo/v4"
)

// defaultTenant is the stub used when multi-tenancy is disabled
// (app.multi_tenancy_enabled = false, the default). It skips host parsing
// and the DB lookup entirely, pinning every request to the tenant every
// existing single-tenant install already has (seeded as id=1 by the
// phase 1 migration) - this keeps the middleware a no-op for every
// deployment that hasn't opted in.
var defaultTenant = &models.Tenant{
	Base:   models.Base{ID: 1},
	Slug:   "default",
	Status: "active",
}

// operatorPathPrefix is exempted from subdomain resolution entirely - the
// Operator API (cmd/operator.go) is inherently cross-tenant, authenticated
// by its own static bearer token rather than a tenant session, and its
// handlers never read the resolved tenant from context.
const operatorPathPrefix = "/api/operator/"

// SlugSuffix returns the subdomain suffix a tenant's slug carries for the
// given [operator].env - "-dev" for "dev", "" for prod (the default,
// empty-string env). A dev deployment appends this to every generated
// tenant URL (cmd/operator.go's tenantRootURL) so its subdomains never
// collide with a prod deployment sharing the same root_domain;
// resolveTenant below strips the same suffix when resolving incoming
// requests. Single source of truth shared by both sides - a mismatch
// would 404 every dev tenant on its own URL.
func SlugSuffix(env string) string {
	if env == "dev" {
		return "-dev"
	}
	return ""
}

// Middleware resolves the tenant for a request and stores it on the echo
// context under models.TenantCtxKey. When enabled is false it always
// resolves to the seeded default tenant without touching the database.
// When enabled, the request's Host header must either carry a subdomain
// of rootDomain (`<slug>.rootDomain`, optionally with the env's
// SlugSuffix - e.g. `<slug>-dev.rootDomain`) or exactly match a tenant's
// tenants.custom_domain, identifying an active tenant, or the request is
// rejected.
func Middleware(core *core.Core, rootDomain string, enabled bool, env string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !enabled || strings.HasPrefix(c.Request().URL.Path, operatorPathPrefix) {
				c.Set(models.TenantCtxKey, defaultTenant)
				return next(c)
			}

			host := c.Request().Host
			if h, _, err := net.SplitHostPort(host); err == nil {
				host = h
			}

			t, err := resolveTenant(core, host, rootDomain, env)
			if err != nil {
				return echo.NewHTTPError(http.StatusNotFound)
			}
			if t.Status != "active" {
				return echo.NewHTTPError(http.StatusServiceUnavailable, "workspace unavailable")
			}

			c.Set(models.TenantCtxKey, &t)
			return next(c)
		}
	}
}

// resolveTenant tries subdomain stripping first (<slug>.rootDomain,
// trimming env's SlugSuffix off the slug if present) - the original,
// still-default resolution path, and the common case for every request
// that isn't under a custom domain - falling back to an exact Host match
// against tenants.custom_domain only when the Host doesn't fit the
// subdomain pattern at all. This ordering means a normal
// <slug>.rootDomain request costs exactly the one DB lookup it always
// has; the second lookup only ever runs for a custom-domain request,
// which previously 404'd here unconditionally.
func resolveTenant(core *core.Core, host, rootDomain, env string) (models.Tenant, error) {
	slug := strings.TrimSuffix(host, "."+rootDomain)
	if slug != host && slug != "" {
		slug = strings.TrimSuffix(slug, SlugSuffix(env))
		return core.GetTenantBySlug(slug)
	}
	return core.GetTenantByCustomDomain(host)
}
