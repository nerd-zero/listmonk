package core

import (
	"net/http"

	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo/v4"
)

// GetTenantBySlug returns the tenant with the given slug. It is a plain,
// unscoped lookup - the `tenants` table itself carries no RLS policy
// (phase 2, v6.5.0) since it's the table that identifies tenants, not one
// scoped by tenant_id.
func (c *Core) GetTenantBySlug(slug string) (models.Tenant, error) {
	out := []models.Tenant{}
	if err := c.q.GetTenantBySlug.Select(&out, slug); err != nil {
		c.log.Printf("error fetching tenant: %v", err)
		return models.Tenant{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "tenant", "error", pqErrMsg(err)))
	}

	if len(out) == 0 {
		return models.Tenant{}, ErrNotFound
	}

	return out[0], nil
}

// GetActiveTenantIDs returns the IDs of all active tenants. Reused by
// boot-time per-tenant loops (e.g. initTxTemplates) - the same underlying
// query internal/manager's scanCampaigns (phase 6) uses via cmd/manager_store.go.
func (c *Core) GetActiveTenantIDs() ([]int, error) {
	var out []int
	if err := c.q.GetActiveTenantIDs.Select(&out); err != nil {
		c.log.Printf("error fetching active tenant IDs: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "tenant", "error", pqErrMsg(err)))
	}

	return out, nil
}
