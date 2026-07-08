// Package operator implements the cross-tenant Operator API: a separate,
// unauthenticated-by-session route group (static bearer token instead of
// user sessions/RLS) for provisioning and managing tenants. See
// docs/design/multi-tenancy.md's "Operator API" section for the full
// design and its accepted v1 limitations (no per-operator identity/audit
// trail - a single shared token authorizes every request).
package operator

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// Middleware checks a static bearer token from config against every
// request. There is no per-operator identity in v1 - see the package
// doc comment for why this is an accepted limitation, not an oversight.
func Middleware(token string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if token == "" {
				return echo.NewHTTPError(http.StatusNotFound)
			}

			got := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")
			if got == "" || subtle.ConstantTimeCompare([]byte(got), []byte(token)) != 1 {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid operator token")
			}

			return next(c)
		}
	}
}
