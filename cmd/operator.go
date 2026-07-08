package main

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/goyesql/v2"
	goyesqlx "github.com/knadh/goyesql/v2/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/listmonk/internal/auth"
	"github.com/knadh/listmonk/internal/core"
	"github.com/knadh/listmonk/internal/tmptokens"
	"github.com/knadh/listmonk/internal/utils"
	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	null "gopkg.in/volatiletech/null.v6"
)

const operatorSetupTokenTTL = 7 * 24 * time.Hour

var reTenantSlug = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

// operatorQueries holds the prepared statements that run exclusively
// against the operator DB connection (see initOperatorDB) - a separate
// pool using a Postgres role with BYPASSRLS, distinct from the
// tenant-app pool. Never share these statements with the main *sqlx.DB.
type operatorQueries struct {
	CreateTenant       *sqlx.Stmt `query:"operator-create-tenant"`
	GetTenant          *sqlx.Stmt `query:"operator-get-tenant"`
	GetTenants         *sqlx.Stmt `query:"operator-get-tenants"`
	UpdateTenantStatus *sqlx.Stmt `query:"operator-update-tenant-status"`
}

// operatorTenant is a tenant row augmented with cross-tenant counts, only
// obtainable via the BYPASSRLS operator connection.
type operatorTenant struct {
	models.Tenant
	UserCount       int `db:"user_count" json:"user_count"`
	SubscriberCount int `db:"subscriber_count" json:"subscriber_count"`
} // @name OperatorTenant

// operatorSetupToken is the payload stored against the one-time setup
// link token returned by CreateTenant (internal/tmptokens, in-memory,
// process-lifetime - same store already used for password-reset/2FA
// tokens, with the same accepted limitation that a restart invalidates
// pending links; see docs/design/multi-tenancy.md's Operator API
// section, "known gap, acceptable for v1").
type operatorSetupToken struct {
	TenantID int
	Email    string
}

// operatorStore bridges the operator DB connection and the normal
// tenant-app Core: tenant CRUD and cross-tenant counts go through the
// BYPASSRLS connection (operatorQueries below); creating the new
// tenant's initial role/user reuses Core.CreateRole/CreateUser on the
// normal pool via WithTenant(newTenantID, ...) - that's a same-tenant
// write using the newly-allocated tenant's own ID, so it needs no
// elevated privilege at all.
type operatorStore struct {
	db          *sqlx.DB
	q           *operatorQueries
	co          *core.Core
	permissions map[string]struct{}
}

// initOperatorDB connects to the same database as the main [db] config
// using a separate Postgres role with BYPASSRLS ([operator].db_user /
// db_password). Returns nil if the Operator API isn't configured
// (empty token or db_user) - it's off by default, an advanced feature
// most single-tenant installs will never enable.
func initOperatorDB(ko *koanf.Koanf) *sqlx.DB {
	if ko.String("operator.token") == "" || ko.String("operator.db_user") == "" {
		return nil
	}

	var c struct {
		Host    string `koanf:"host"`
		Port    int    `koanf:"port"`
		DBName  string `koanf:"database"`
		SSLMode string `koanf:"ssl_mode"`
		Params  string `koanf:"params"`
	}
	if err := ko.Unmarshal("db", &c); err != nil {
		lo.Fatalf("error loading db config for operator connection: %v", err)
	}

	fields := map[string]string{
		"host":     c.Host,
		"port":     strconv.Itoa(c.Port),
		"user":     ko.String("operator.db_user"),
		"password": ko.String("operator.db_password"),
		"dbname":   c.DBName,
		"sslmode":  c.SSLMode,
	}
	if c.Port == 0 {
		delete(fields, "port")
	}

	var parts []string
	for k, v := range fields {
		if v == "" {
			continue
		}
		parts = append(parts, k+"="+v)
	}
	if c.Params != "" {
		parts = append(parts, c.Params)
	}

	db, err := sqlx.Connect("postgres", strings.Join(parts, " "))
	if err != nil {
		lo.Fatalf("error connecting to operator DB: %v", err)
	}

	return db
}

// newOperatorStoreIfEnabled connects to the operator DB (if configured)
// and prepares its queries against the same parsed SQL map every other
// query set is prepared from (queries/operator.sql is loaded alongside
// the rest). Returns nil if the Operator API isn't configured.
func newOperatorStoreIfEnabled(qMap goyesql.Queries, co *core.Core, permissions map[string]struct{}) *operatorStore {
	db := initOperatorDB(ko)
	if db == nil {
		return nil
	}

	var q operatorQueries
	if err := goyesqlx.ScanToStruct(&q, qMap, db); err != nil {
		lo.Fatalf("error preparing operator SQL queries: %v", err)
	}

	lo.Println("operator API enabled")
	return &operatorStore{db: db, q: &q, co: co, permissions: permissions}
}

func (s *operatorStore) GetTenants() ([]operatorTenant, error) {
	out := []operatorTenant{}
	if err := s.q.GetTenants.Select(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *operatorStore) GetTenant(id int) (operatorTenant, error) {
	var out operatorTenant
	if err := s.q.GetTenant.Get(&out, id); err != nil {
		return out, err
	}
	return out, nil
}

func (s *operatorStore) UpdateTenantStatus(id int, status string) (models.Tenant, error) {
	var out models.Tenant
	if err := s.q.UpdateTenantStatus.Get(&out, id, status); err != nil {
		return out, err
	}
	return out, nil
}

// CreateTenant creates a tenant row, then - via the normal tenant-app
// Core, not the operator connection - a "Super Admin" role with every
// permission and an initial admin user for it (PasswordLogin disabled,
// no password set yet). Returns the tenant and a one-time setup token
// the caller is responsible for delivering to the new admin (v1
// placeholder: returned directly in the API response - see the design
// doc's Operator API section for why e-mailing it isn't possible yet,
// a brand new tenant has no SMTP config of its own).
func (s *operatorStore) CreateTenant(ctx context.Context, slug, name, adminUsername, adminEmail string) (models.Tenant, string, error) {
	var t models.Tenant
	if err := s.q.CreateTenant.Get(&t, slug, name); err != nil {
		return t, "", err
	}

	r := auth.Role{
		Type: auth.RoleTypeUser,
		Name: null.NewString("Super Admin", true),
	}
	for p := range s.permissions {
		r.Permissions = append(r.Permissions, p)
	}

	newRole, err := s.co.CreateRole(ctx, t.ID, r)
	if err != nil {
		return t, "", err
	}

	u := auth.User{
		Type:          auth.UserTypeUser,
		HasPassword:   false,
		PasswordLogin: false,
		Username:      adminUsername,
		Name:          adminUsername,
		Email:         null.NewString(adminEmail, true),
		UserRoleID:    newRole.ID,
		Status:        auth.UserStatusEnabled,
	}
	if _, err := s.co.CreateUser(ctx, t.ID, u); err != nil {
		return t, "", err
	}

	token, err := generateRandomString(tmpAuthTokenLen)
	if err != nil {
		return t, "", err
	}
	tmptokens.Set(token, operatorSetupTokenTTL, operatorSetupToken{TenantID: t.ID, Email: adminEmail})

	return t, token, nil
}

// CreateSetupLink issues a fresh one-time setup token for an existing
// tenant admin, without creating a new tenant/user. Needed because
// internal/tmptokens is in-memory and process-lifetime only (see its own
// docs) - a setup link issued by CreateTenant is lost on every app
// restart, and there was previously no way to recover from that short of
// recreating the tenant from scratch. Verifies the user actually exists
// for this tenant/email first so this can't be used to probe for emails
// across tenants.
func (s *operatorStore) CreateSetupLink(ctx context.Context, tenantID int, email string) (string, error) {
	if _, err := s.co.GetUser(ctx, tenantID, 0, "", email); err != nil {
		return "", err
	}

	token, err := generateRandomString(tmpAuthTokenLen)
	if err != nil {
		return "", err
	}
	tmptokens.Set(token, operatorSetupTokenTTL, operatorSetupToken{TenantID: tenantID, Email: email})

	return token, nil
}

// operatorTenantReq is the request body for CreateTenant.
type operatorTenantReq struct {
	Slug          string `json:"slug"`
	Name          string `json:"name"`
	AdminUsername string `json:"admin_username"`
	AdminEmail    string `json:"admin_email"`
} // @name OperatorCreateTenantReq

// operatorCreateTenantResp is the response body for CreateOperatorTenant.
type operatorCreateTenantResp struct {
	Tenant     models.Tenant `json:"tenant"`
	SetupToken string        `json:"setup_token"`
	SetupURL   string        `json:"setup_url,omitempty"`
} // @name OperatorCreateTenantResp

// operatorSetupLinkReq is the request body for CreateOperatorSetupLink.
type operatorSetupLinkReq struct {
	AdminEmail string `json:"admin_email"`
} // @name OperatorCreateSetupLinkReq

// operatorSetupLinkResp is the response body for CreateOperatorSetupLink.
type operatorSetupLinkResp struct {
	SetupToken string `json:"setup_token"`
	SetupURL   string `json:"setup_url,omitempty"`
} // @name OperatorCreateSetupLinkResp

// operatorUpdateStatusReq is the request body for UpdateOperatorTenantStatus.
type operatorUpdateStatusReq struct {
	Status string `json:"status"`
} // @name OperatorUpdateTenantStatusReq

// ListOperatorTenants returns every tenant with basic cross-tenant counts.
//
//	@ID			listOperatorTenants
//	@Summary		List tenants (Operator API)
//	@Description	Fork-only, off by default (see [operator] config). Requires the Operator API bearer token, not a normal session/API-user token.
//	@Tags			operator
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	[]operatorTenant
//	@Failure		401	{object}	echo.HTTPError
//	@Failure		500	{object}	echo.HTTPError
//	@Router			/api/operator/tenants [get]
func (a *App) ListOperatorTenants(c echo.Context) error {
	out, err := a.operator.GetTenants()
	if err != nil {
		a.log.Printf("error fetching tenants: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "error fetching tenants")
	}
	return c.JSON(http.StatusOK, okResp{out})
}

// GetOperatorTenant returns a single tenant with basic cross-tenant counts.
//
//	@ID			getOperatorTenant
//	@Summary		Get a tenant (Operator API)
//	@Description	Fork-only, off by default (see [operator] config). Requires the Operator API bearer token, not a normal session/API-user token.
//	@Tags			operator
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Tenant ID"
//	@Success		200	{object}	operatorTenant
//	@Failure		401	{object}	echo.HTTPError
//	@Failure		404	{object}	echo.HTTPError
//	@Router			/api/operator/tenants/{id} [get]
func (a *App) GetOperatorTenant(c echo.Context) error {
	id := getID(c)
	out, err := a.operator.GetTenant(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "tenant not found")
	}
	return c.JSON(http.StatusOK, okResp{out})
}

// CreateOperatorTenant provisions a new tenant and its initial admin user.
//
//	@ID			createOperatorTenant
//	@Summary		Create a tenant (Operator API)
//	@Description	Fork-only, off by default (see [operator] config). Requires the Operator API bearer token, not a normal session/API-user token. Creates the tenant plus a passwordless initial admin user; the returned setup_url/setup_token is a one-time link the tenant's actual admin uses to set their own password - the operator never sets or sees it.
//	@Tags			operator
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			tenant	body		operatorTenantReq	true	"Tenant to create"
//	@Success		200	{object}	operatorCreateTenantResp
//	@Failure		400	{object}	echo.HTTPError
//	@Failure		401	{object}	echo.HTTPError
//	@Failure		409	{object}	echo.HTTPError	"Slug already in use"
//	@Failure		500	{object}	echo.HTTPError
//	@Router			/api/operator/tenants [post]
func (a *App) CreateOperatorTenant(c echo.Context) error {
	var req operatorTenantReq
	if err := c.Bind(&req); err != nil {
		return err
	}

	req.Slug = strings.ToLower(strings.TrimSpace(req.Slug))
	if !reTenantSlug.MatchString(req.Slug) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid slug: use lowercase letters, numbers, hyphens")
	}
	if !strHasLen(req.Name, 1, stdInputMaxLen) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid name")
	}
	if !strHasLen(req.AdminUsername, 3, stdInputMaxLen) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid admin_username")
	}
	if !utils.ValidateEmail(req.AdminEmail) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid admin_email")
	}

	tenant, token, err := a.operator.CreateTenant(c.Request().Context(), req.Slug, req.Name, req.AdminUsername, req.AdminEmail)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Constraint == "tenants_slug_key" {
			return echo.NewHTTPError(http.StatusConflict, "a tenant with this slug already exists")
		}
		a.log.Printf("error creating tenant: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "error creating tenant")
	}

	return c.JSON(http.StatusOK, okResp{operatorCreateTenantResp{tenant, token, a.operatorSetupURL(tenant.Slug, token)}})
}

// operatorSetupURL builds the one-time setup link for a tenant's admin,
// shared by CreateOperatorTenant and CreateOperatorSetupLink. Empty if
// app.root_domain isn't configured or app.root_url has no scheme.
func (a *App) operatorSetupURL(tenantSlug, token string) string {
	if a.cfg.RootDomain == "" {
		return ""
	}
	u, err := url.Parse(a.urlCfg.RootURL)
	if err != nil || u.Scheme == "" {
		return ""
	}
	return u.Scheme + "://" + tenantSlug + "." + a.cfg.RootDomain + "/admin/operator-setup?token=" + token
}

// CreateOperatorSetupLink issues a fresh one-time setup link for an
// existing tenant's admin. Needed because setup tokens live only in
// internal/tmptokens' in-memory store - restarting the app invalidates
// every pending link issued by CreateOperatorTenant, and until this
// endpoint existed the only recovery was recreating the tenant from
// scratch (which fails outright since the slug is already taken).
//
//	@ID			createOperatorSetupLink
//	@Summary		Reissue a tenant admin's setup link (Operator API)
//	@Description	Fork-only, off by default (see [operator] config). Requires the Operator API bearer token. Use when a prior setup_url from CreateOperatorTenant expired or was lost - e.g. every pending link is invalidated on app restart, since setup tokens are held in memory only.
//	@Tags			operator
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int							true	"Tenant ID"
//	@Param			email	body		operatorSetupLinkReq	true	"Existing admin's email"
//	@Success		200	{object}	operatorSetupLinkResp
//	@Failure		400	{object}	echo.HTTPError
//	@Failure		401	{object}	echo.HTTPError
//	@Failure		404	{object}	echo.HTTPError	"Tenant or admin email not found"
//	@Router			/api/operator/tenants/{id}/setup-link [post]
func (a *App) CreateOperatorSetupLink(c echo.Context) error {
	id := getID(c)

	var req operatorSetupLinkReq
	if err := c.Bind(&req); err != nil {
		return err
	}
	if !utils.ValidateEmail(req.AdminEmail) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid admin_email")
	}

	tenant, err := a.operator.GetTenant(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "tenant not found")
	}

	token, err := a.operator.CreateSetupLink(c.Request().Context(), id, req.AdminEmail)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "no admin with that email found for this tenant")
	}

	return c.JSON(http.StatusOK, okResp{operatorSetupLinkResp{token, a.operatorSetupURL(tenant.Slug, token)}})
}

// UpdateOperatorTenantStatus updates a tenant's status (active/suspended/disabled).
//
//	@ID			updateOperatorTenantStatus
//	@Summary		Update a tenant's status (Operator API)
//	@Description	Fork-only, off by default (see [operator] config). Requires the Operator API bearer token, not a normal session/API-user token.
//	@Tags			operator
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int							true	"Tenant ID"
//	@Param			status	body		operatorUpdateStatusReq	true	"New status"
//	@Success		200	{object}	models.Tenant
//	@Failure		400	{object}	echo.HTTPError
//	@Failure		401	{object}	echo.HTTPError
//	@Failure		500	{object}	echo.HTTPError
//	@Router			/api/operator/tenants/{id}/status [put]
func (a *App) UpdateOperatorTenantStatus(c echo.Context) error {
	id := getID(c)

	var req operatorUpdateStatusReq
	if err := c.Bind(&req); err != nil {
		return err
	}
	if req.Status != "active" && req.Status != "suspended" && req.Status != "disabled" {
		return echo.NewHTTPError(http.StatusBadRequest, "status must be one of active, suspended, disabled")
	}

	out, err := a.operator.UpdateTenantStatus(id, req.Status)
	if err != nil {
		a.log.Printf("error updating tenant status: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "error updating tenant status")
	}

	return c.JSON(http.StatusOK, okResp{out})
}

// OperatorSetupPage renders (GET) and processes (POST) the one-time
// setup link an operator-provisioned tenant's initial admin uses to set
// their password. Public - authorized entirely by possessing the
// one-time token, same trust model as the existing password-reset flow.
func (a *App) OperatorSetupPage(c echo.Context) error {
	token := strings.TrimSpace(c.QueryParam("token"))
	if c.Request().Method == http.MethodPost {
		token = strings.TrimSpace(c.FormValue("token"))
	}

	data, err := tmptokens.Check(token)
	if err != nil {
		return c.Render(http.StatusBadRequest, tplMessage, makeMsgTpl(a.i18n.T("users.resetPassword"), "", a.i18n.T("users.invalidResetLink")))
	}
	payload, ok := data.(operatorSetupToken)
	if !ok {
		return c.Render(http.StatusBadRequest, tplMessage, makeMsgTpl(a.i18n.T("users.resetPassword"), "", a.i18n.T("users.invalidResetLink")))
	}

	if c.Request().Method == http.MethodPost {
		return a.doOperatorSetup(c, token, payload)
	}

	return c.Render(http.StatusOK, "admin-operator-setup", resetPasswordTpl{
		Title: a.i18n.T("users.resetPassword"),
		Token: token,
		Email: payload.Email,
	})
}

func (a *App) doOperatorSetup(c echo.Context, token string, payload operatorSetupToken) error {
	var (
		password  = c.FormValue("password")
		password2 = c.FormValue("password2")
	)

	if !strHasLen(password, 8, stdInputMaxLen) {
		return c.Render(http.StatusBadRequest, "admin-operator-setup", resetPasswordTpl{
			Title: a.i18n.T("users.resetPassword"), Token: token, Email: payload.Email,
			Error: a.i18n.Ts("globals.messages.invalidFields", "name", "password"),
		})
	}
	if password != password2 {
		return c.Render(http.StatusBadRequest, "admin-operator-setup", resetPasswordTpl{
			Title: a.i18n.T("users.resetPassword"), Token: token, Email: payload.Email,
			Error: a.i18n.T("users.passwordMismatch"),
		})
	}

	// Consume the token so it can't be reused.
	if _, err := tmptokens.Get(token); err != nil {
		return c.Render(http.StatusBadRequest, tplMessage, makeMsgTpl(a.i18n.T("users.resetPassword"), "", a.i18n.T("users.invalidResetLink")))
	}

	user, err := a.core.GetUser(c.Request().Context(), payload.TenantID, 0, "", payload.Email)
	if err != nil {
		return c.Render(http.StatusBadRequest, tplMessage, makeMsgTpl(a.i18n.T("users.resetPassword"), "", a.i18n.T("users.invalidResetLink")))
	}

	// UpdateUserProfile (not UpdateUser) - same as the analogous
	// forgot-password flow (cmd/auth.go's doResetPassword). UpdateUser is
	// for admin-editing-another-user and re-runs the last-Super-Admin
	// guard against role_id/status; GetUser zeroes UserRoleID after
	// copying it into UserRole.ID (setupUserFields, for JSON shaping), so
	// passing it straight back to UpdateUser sent role_id=0 - the query's
	// "unchanged" sentinel for a normal edit, but one the guard's OR
	// condition didn't recognize as "unchanged," so it always treated a
	// lone tenant admin's own password-setup as "reassigning away from
	// Super Admin" and rejected it. UpdateUserProfile only ever touches
	// name/email/password, sidestepping that guard entirely - correct,
	// since setting your own password isn't a role/status change.
	user.Password = null.NewString(password, true)
	user.PasswordLogin = true
	if _, err := a.core.UpdateUserProfile(c.Request().Context(), payload.TenantID, user.ID, user); err != nil {
		a.log.Printf("error completing operator setup for user_id=%d: %v", user.ID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, a.i18n.T("globals.messages.internalError"))
	}

	// Log the user in directly and land on the dashboard, rather than a
	// dead-end "Done" message with no way to continue - same UX as the
	// analogous forgot-password flow (cmd/auth.go's doResetPassword).
	if err := a.auth.SaveSession(user, "", c); err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, uriAdmin)
}
