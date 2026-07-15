// Package cloudflareclient is a thin, hand-rolled HTTP client for
// Cloudflare's Custom Hostnames API ("Cloudflare for SaaS") -- same "no
// SDK, small surface" reasoning already applied to internal/postmarkclient
// and internal/operatorclient rather than pulling in cloudflare-go. Only
// what internal/provisioning's AddCustomDomain/GetCustomDomain/
// DeleteCustomDomain need: create, get, and delete a Custom Hostname on
// one zone. See docs/custom-domains.md.
//
// The org's own DNS never has to be on Cloudflare: they CNAME their
// domain at our own fallback-origin hostname (CLOUDFLARE_FALLBACK_ORIGIN,
// itself proxied through our zone), which works the same regardless of
// their registrar/DNS host. This client only ever talks to our zone.
package cloudflareclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultBaseURL = "https://api.cloudflare.com/client/v4"

// Client talks to one Cloudflare zone's Custom Hostnames API using a
// scoped API token (Zone:SSL and Certificates:Edit + Zone:Custom
// Hostname:Edit -- see docs/custom-domains.md). Not tied to any one
// tenant -- every call takes the hostname/ID it operates on explicitly.
type Client struct {
	baseURL  string
	apiToken string
	zoneID   string
	http     *http.Client
}

func New(apiToken, zoneID string) *Client {
	return &Client{
		baseURL:  defaultBaseURL,
		apiToken: apiToken,
		zoneID:   zoneID,
		http:     &http.Client{Timeout: 15 * time.Second},
	}
}

// OwnershipVerification is the TXT record Cloudflare requires published
// before it'll issue a cert for a Custom Hostname -- proves the org
// controls Hostname, independent of (and required before) the CNAME
// pointing it at our fallback origin.
type OwnershipVerification struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

// SSL mirrors the subset of a Custom Hostname's nested "ssl" object this
// client needs. Status moves through "pending_validation" ->
// "pending_issuance" -> "pending_deployment" -> "active" as Cloudflare
// notices the org's DNS records published and issues + deploys a cert --
// IsActive checks for the last of these. TxtName/TxtValue are the CA's
// own domain-control-validation TXT record (e.g. served under
// "_acme-challenge.<hostname>" by Google Trust Services) -- separate from,
// and in addition to, CustomHostname.OwnershipVerification below. Both
// must be published for SSL.Status to ever reach "active".
type SSL struct {
	Status   string `json:"status"`
	TxtName  string `json:"txt_name"`
	TxtValue string `json:"txt_value"`
}

// CustomHostname mirrors the subset of Cloudflare's Custom Hostname
// resource this client needs. Before SSL.Status ever reaches "active",
// the org must publish three DNS records, none of which require them to
// use Cloudflare themselves: a CNAME pointing Hostname at our zone's
// configured fallback origin (a plain config value, CLOUDFLARE_FALLBACK_ORIGIN --
// not part of this API response, since it's a zone-level setting, not a
// per-hostname one), the OwnershipVerification TXT record below (proves
// hostname ownership to Cloudflare itself, checked first), and the
// SSL.TxtName/TxtValue TXT record (the issuing CA's own DCV challenge,
// checked once ownership passes -- these are two genuinely different
// records, easy to conflate since both are TXT).
type CustomHostname struct {
	ID                    string                `json:"id"`
	Hostname              string                `json:"hostname"`
	SSL                   SSL                   `json:"ssl"`
	OwnershipVerification OwnershipVerification `json:"ownership_verification"`
}

// IsActive reports whether Cloudflare has finished domain control
// validation and issued + deployed a certificate for this hostname -- the
// point at which internal/provisioning.GetCustomDomain flips the tenant's
// app.root_url to it (docs/custom-domains.md's proposed flow, steps 4-5).
func (h CustomHostname) IsActive() bool {
	return h.SSL.Status == "active"
}

// CreateCustomHostname registers hostname as a new Custom Hostname on
// this zone, using TXT-based domain control validation -- works
// regardless of whether the org's CNAME is in place yet, unlike the
// alternative "http" validation method. The response's
// OwnershipVerification record is what the org must publish before a
// cert is issued.
func (c *Client) CreateCustomHostname(ctx context.Context, hostname string) (CustomHostname, error) {
	body := map[string]any{
		"hostname": hostname,
		"ssl": map[string]any{
			"method": "txt",
			"type":   "dv",
		},
	}
	return doJSON[CustomHostname](ctx, c, http.MethodPost, "/zones/"+c.zoneID+"/custom_hostnames", body)
}

// GetCustomHostname fetches a Custom Hostname's current live state --
// used to notice SSL.Status flipping to "active" once the org has
// published both required DNS records.
func (c *Client) GetCustomHostname(ctx context.Context, id string) (CustomHostname, error) {
	return doJSON[CustomHostname](ctx, c, http.MethodGet, "/zones/"+c.zoneID+"/custom_hostnames/"+id, nil)
}

// DeleteCustomHostname permanently removes a Custom Hostname --
// irreversible; Cloudflare stops proxying/terminating TLS for it
// immediately.
func (c *Client) DeleteCustomHostname(ctx context.Context, id string) error {
	_, err := doJSON[struct{}](ctx, c, http.MethodDelete, "/zones/"+c.zoneID+"/custom_hostnames/"+id, nil)
	return err
}

// IsNotFound reports whether err is Cloudflare's response for a Custom
// Hostname ID that no longer exists -- lets a delete be treated as
// already-done rather than a failure when retried.
func IsNotFound(err error) bool {
	var ae *APIError
	return errors.As(err, &ae) && ae.StatusCode == http.StatusNotFound
}

// APIError is returned whenever Cloudflare's envelope carries
// "success": false -- the shape every endpoint uses
// (developers.cloudflare.com/api).
type APIError struct {
	StatusCode int
	Code       int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("cloudflare API: %d (code %d) %s", e.StatusCode, e.Code, e.Message)
}

// envelope mirrors Cloudflare's {"success", "errors", "result"} wrapper
// every API response uses, success or not.
type envelope[T any] struct {
	Success bool `json:"success"`
	Errors  []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
	Result T `json:"result"`
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
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return zero, fmt.Errorf("cloudflare API request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return zero, fmt.Errorf("read response body: %w", err)
	}

	var env envelope[T]
	if err := json.Unmarshal(respBody, &env); err != nil {
		return zero, fmt.Errorf("decode response body: %w", err)
	}

	if !env.Success || resp.StatusCode >= 300 {
		msg := string(respBody)
		code := 0
		if len(env.Errors) > 0 {
			msg = env.Errors[0].Message
			code = env.Errors[0].Code
		}
		return zero, &APIError{StatusCode: resp.StatusCode, Code: code, Message: msg}
	}

	return env.Result, nil
}
