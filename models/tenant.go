package models

// TenantCtxKey is the echo.Context key the tenant-resolution middleware
// (internal/tenant) stores the resolved *Tenant under, and internal/auth
// reads to cross-check against the authenticated user's TenantID. It lives
// here, not in internal/tenant itself, because internal/core already
// imports internal/auth - internal/auth importing internal/tenant (which
// must import internal/core for the DB lookup) would cycle.
const TenantCtxKey = "tenant"

// Tenant represents a single tenant (organization) in a multi-tenant
// listmonk deployment.
type Tenant struct {
	Base

	Slug   string `db:"slug" json:"slug"`
	Name   string `db:"name" json:"name"`
	Status string `db:"status" json:"status"`
}
