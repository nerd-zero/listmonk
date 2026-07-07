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

// Middleware resolves the tenant for a request and stores it on the echo
// context under models.TenantCtxKey. When enabled is false it always
// resolves to the seeded default tenant without touching the database.
// When enabled, it requires the request's Host header to carry a
// subdomain of rootDomain (`<slug>.rootDomain`) identifying an active
// tenant, or the request is rejected.
func Middleware(core *core.Core, rootDomain string, enabled bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !enabled {
				c.Set(models.TenantCtxKey, defaultTenant)
				return next(c)
			}

			host := c.Request().Host
			if h, _, err := net.SplitHostPort(host); err == nil {
				host = h
			}

			slug := strings.TrimSuffix(host, "."+rootDomain)
			if slug == host || slug == "" {
				return echo.NewHTTPError(http.StatusNotFound)
			}

			t, err := core.GetTenantBySlug(slug)
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
