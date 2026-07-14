// Package operatorclient is a thin, hand-rolled HTTP client for the
// listmonk fork's cross-tenant Operator API (/api/operator/*, see
// docs/design/multi-tenancy.md and cmd/operator.go in that repo). No SDK
// exists for this since it's a fork-only surface, not something upstream
// ships -- matching the project's existing "thin hand-rolled client, no
// SDK" pattern already chosen for Postmark in docs/plan.md.
package operatorclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Client talks to one listmonk fork's Operator API using its static
// bearer token. Not tied to any one tenant -- every call is cross-tenant.
type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func New(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		http:    &http.Client{Timeout: 15 * time.Second},
	}
}

// Tenant mirrors the listmonk fork's models.Tenant.
type Tenant struct {
	ID        int       `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	Status    string    `json:"status"` // active | suspended | disabled
}

// TenantWithCounts mirrors cmd/operator.go's operatorTenant -- a Tenant
// plus cross-tenant counts only obtainable via the fork's BYPASSRLS
// operator DB connection.
type TenantWithCounts struct {
	Tenant
	UserCount       int `json:"user_count"`
	SubscriberCount int `json:"subscriber_count"`
}

// CreateTenantParams mirrors cmd/operator.go's operatorTenantReq.
// OrganizationID is optional (0 means none), mirroring the fork's own
// "organizationID <= 0 means none" convention.
type CreateTenantParams struct {
	Slug           string `json:"slug"`
	Name           string `json:"name"`
	AdminUsername  string `json:"admin_username"`
	AdminEmail     string `json:"admin_email"`
	OrganizationID int    `json:"organization_id,omitempty"`
}

// Organization mirrors the listmonk fork's models.Organization -- a purely
// cross-tenant grouping construct (never RLS-scoped, never resolved
// per-request) that tenants can optionally belong to.
type Organization struct {
	ID        int       `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
}

// OrganizationWithCount mirrors cmd/operator.go's operatorOrganization.
type OrganizationWithCount struct {
	Organization
	TenantCount int `json:"tenant_count"`
}

// OrganizationDetail mirrors cmd/operator.go's operatorOrganizationResp.
type OrganizationDetail struct {
	OrganizationWithCount
	Tenants []TenantWithCounts `json:"tenants"`
}

// CreateTenantResult mirrors cmd/operator.go's operatorCreateTenantResp.
// SetupURL is the one-time link the new admin uses to set their password --
// it can't be emailed by listmonk itself (a brand-new tenant has no SMTP
// config yet), so the caller is responsible for delivering it. The token
// behind it lives in the fork's memory only and is lost on its restart.
type CreateTenantResult struct {
	Tenant     Tenant `json:"tenant"`
	SetupToken string `json:"setup_token"`
	SetupURL   string `json:"setup_url"`
}

// SetupLinkResult mirrors cmd/operator.go's operatorSetupLinkResp.
type SetupLinkResult struct {
	SetupToken string `json:"setup_token"`
	SetupURL   string `json:"setup_url"`
}

// APIError is returned for any non-2xx response, carrying the status code
// and the message the fork's echo.HTTPError handler serializes as
// {"message": "..."}.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("listmonk operator API: %d %s", e.StatusCode, e.Message)
}

// IsConflict reports whether err is a 409 -- the fork's response to a
// duplicate tenant slug (cmd/operator.go's CreateOperatorTenant). Callers
// use this to turn a slug collision into a clean validation error instead
// of a generic failure.
func IsConflict(err error) bool {
	var ae *APIError
	return errors.As(err, &ae) && ae.StatusCode == http.StatusConflict
}

// IsNotFound reports whether err is a 404.
func IsNotFound(err error) bool {
	var ae *APIError
	return errors.As(err, &ae) && ae.StatusCode == http.StatusNotFound
}

func (c *Client) CreateTenant(ctx context.Context, p CreateTenantParams) (CreateTenantResult, error) {
	return doJSON[CreateTenantResult](ctx, c, http.MethodPost, "/api/operator/tenants", p)
}

func (c *Client) CreateOrganization(ctx context.Context, name string) (Organization, error) {
	return doJSON[Organization](ctx, c, http.MethodPost, "/api/operator/organizations", map[string]string{"name": name})
}

func (c *Client) ListOrganizations(ctx context.Context) ([]OrganizationWithCount, error) {
	return doJSON[[]OrganizationWithCount](ctx, c, http.MethodGet, "/api/operator/organizations", nil)
}

func (c *Client) GetOrganization(ctx context.Context, id int) (OrganizationDetail, error) {
	path := "/api/operator/organizations/" + strconv.Itoa(id)
	return doJSON[OrganizationDetail](ctx, c, http.MethodGet, path, nil)
}

func (c *Client) ListTenants(ctx context.Context) ([]TenantWithCounts, error) {
	return doJSON[[]TenantWithCounts](ctx, c, http.MethodGet, "/api/operator/tenants", nil)
}

func (c *Client) GetTenant(ctx context.Context, id int) (TenantWithCounts, error) {
	path := "/api/operator/tenants/" + strconv.Itoa(id)
	return doJSON[TenantWithCounts](ctx, c, http.MethodGet, path, nil)
}

func (c *Client) UpdateTenantStatus(ctx context.Context, id int, status string) (Tenant, error) {
	path := "/api/operator/tenants/" + strconv.Itoa(id) + "/status"
	return doJSON[Tenant](ctx, c, http.MethodPut, path, map[string]string{"status": status})
}

// CreateSetupLink reissues a one-time setup link for an existing tenant
// admin -- backs the dashboard's "resend setup link" action, needed
// because the original link's token is lost on every listmonk restart.
func (c *Client) CreateSetupLink(ctx context.Context, id int, adminEmail string) (SetupLinkResult, error) {
	path := "/api/operator/tenants/" + strconv.Itoa(id) + "/setup-link"
	return doJSON[SetupLinkResult](ctx, c, http.MethodPost, path, map[string]string{"admin_email": adminEmail})
}

// SMTPEntry mirrors the listmonk fork's cmd/operator.go operatorSMTPEntry --
// one entry of models.Settings' own SMTP field. listmonk has no
// provider-specific knowledge: whoever calls SetTenantSMTP (here,
// internal/provisioning after internal/postmarkclient creates the actual
// Postmark server) owns creating the real credentials; this struct only
// carries them across the wire.
type SMTPEntry struct {
	Name          string              `json:"name"`
	Enabled       bool                `json:"enabled"`
	Host          string              `json:"host"`
	HelloHostname string              `json:"hello_hostname"`
	Port          int                 `json:"port"`
	AuthProtocol  string              `json:"auth_protocol"`
	Username      string              `json:"username"`
	Password      string              `json:"password"`
	EmailHeaders  []map[string]string `json:"email_headers"`
	MaxConns      int                 `json:"max_conns"`
	MaxMsgRetries int                 `json:"max_msg_retries"`
	MsgRetryDelay string              `json:"msg_retry_delay"`
	IdleTimeout   string              `json:"idle_timeout"`
	WaitTimeout   string              `json:"wait_timeout"`
	TLSType       string              `json:"tls_type"`
	TLSSkipVerify bool                `json:"tls_skip_verify"`
	FromAddresses []string            `json:"from_addresses"`
}

// SetTenantSMTP replaces a tenant's SMTP settings with a single entry --
// the fork itself assigns a fresh UUID and stores it verbatim, no merge
// with whatever placeholder entries seeded the tenant at creation.
func (c *Client) SetTenantSMTP(ctx context.Context, tenantID int, entry SMTPEntry) error {
	path := "/api/operator/tenants/" + strconv.Itoa(tenantID) + "/smtp"
	_, err := doJSON[bool](ctx, c, http.MethodPut, path, entry)
	return err
}

// envelope mirrors the fork's okResp{data} wrapper (cmd's shared response
// helper) that every successful Operator API response is nested under.
type envelope[T any] struct {
	Data T `json:"data"`
}

func doJSON[T any](ctx context.Context, c *Client, method, path string, body any) (T, error) {
	var zero T

	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return zero, fmt.Errorf("encode request body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return zero, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return zero, fmt.Errorf("listmonk operator API request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return zero, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode >= 300 {
		var e struct {
			Message string `json:"message"`
		}
		_ = json.Unmarshal(respBody, &e)
		if e.Message == "" {
			e.Message = string(respBody)
		}
		return zero, &APIError{StatusCode: resp.StatusCode, Message: e.Message}
	}

	if len(respBody) == 0 {
		return zero, nil
	}

	var env envelope[T]
	if err := json.Unmarshal(respBody, &env); err != nil {
		return zero, fmt.Errorf("decode response body: %w", err)
	}
	return env.Data, nil
}
