// Package postmarkclient is a thin, hand-rolled HTTP client for Postmark's
// account-level Account API (https://postmarkapp.com/developer/api/servers-api,
// .../domains-api, .../signatures-api) -- same "no SDK, small surface"
// reasoning already applied to internal/operatorclient. One Postmark server
// per tenant is created up front by internal/provisioning.CreateInstance;
// the tenant's own sending domain or sender signature is added later, by
// the org itself, via internal/provisioning's AddSenderDomain/
// AddSenderSignature. See docs/plan.md's Postmark section.
package postmarkclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const defaultBaseURL = "https://api.postmarkapp.com"

// Client authenticates with a Postmark account-level API token -- not tied
// to any one tenant's server, since creating a server is exactly what this
// client is for.
type Client struct {
	baseURL      string
	accountToken string
	http         *http.Client
}

func New(accountToken string) *Client {
	return &Client{
		baseURL:      defaultBaseURL,
		accountToken: accountToken,
		http:         &http.Client{Timeout: 15 * time.Second},
	}
}

// Server mirrors the subset of Postmark's Server resource this client
// needs. ApiTokens holds up to 3 tokens on creation; the first one doubles
// as the SMTP username and password (postmarkapp.com/support/article/811).
type Server struct {
	ID        int      `json:"ID"`
	Name      string   `json:"Name"`
	ApiTokens []string `json:"ApiTokens"`
}

// CreateServer creates a new Postmark server -- listnun's "one Postmark
// server per tenant" unit. name has no uniqueness requirement on Postmark's
// side, unlike listmonk tenant slugs.
func (c *Client) CreateServer(ctx context.Context, name string) (Server, error) {
	body := map[string]any{
		"Name":             name,
		"SmtpApiActivated": true,
	}
	return doJSON[Server](ctx, c, http.MethodPost, "/servers", body)
}

// Domain mirrors the subset of Postmark's Domain resource this client
// needs to publish the DKIM record a customer (or listnun's own DNS, once
// automated) must add before mail from this domain is trusted. Return-Path
// (bounce domain) setup is a later step -- not configured on creation.
// DKIMVerified flips to true once Postmark's own periodic DNS check (or an
// explicit PUT /domains/{id}/verifyDkim, not yet called by this client)
// finds the record published -- see GetDomain.
type Domain struct {
	ID            int    `json:"ID"`
	Name          string `json:"Name"`
	DKIMHost      string `json:"DKIMHost"`
	DKIMTextValue string `json:"DKIMTextValue"`
	DKIMVerified  bool   `json:"DKIMVerified"`
}

// CreateDomain registers a new sending domain and returns the DKIM record
// to publish. Verification (PUT /domains/{id}/verifyDkim) is a separate,
// not-yet-automated step -- see docs/plan.md's Postmark section.
func (c *Client) CreateDomain(ctx context.Context, name string) (Domain, error) {
	return doJSON[Domain](ctx, c, http.MethodPost, "/domains", map[string]string{"Name": name})
}

// VerifyDKIM actively re-checks DNS for a domain's DKIM record right now,
// rather than waiting for Postmark's own periodic check -- what makes
// noticing a customer's just-published record responsive instead of
// eventually-consistent on Postmark's own schedule.
func (c *Client) VerifyDKIM(ctx context.Context, id int) (Domain, error) {
	return doJSON[Domain](ctx, c, http.MethodPut, "/domains/"+strconv.Itoa(id)+"/verifyDkim", nil)
}

// DeleteDomain permanently removes a sending domain -- irreversible.
func (c *Client) DeleteDomain(ctx context.Context, id int) error {
	_, err := doJSON[struct{}](ctx, c, http.MethodDelete, "/domains/"+strconv.Itoa(id), nil)
	return err
}

// SenderSignature mirrors the subset of Postmark's Sender Signature
// resource this client needs -- the alternative to a full Domain for a
// customer who wants to send from one address (e.g. hello@theirdomain.com)
// without handing over DNS control. Confirmed is always false on creation:
// Postmark emails a confirmation link directly to EmailAddress, and
// there's no API call to complete that step -- only the customer clicking
// the link in their inbox does.
type SenderSignature struct {
	ID           int    `json:"ID"`
	EmailAddress string `json:"EmailAddress"`
	Confirmed    bool   `json:"Confirmed"`
}

// CreateSenderSignature registers a new sender signature. Postmark sends
// the confirmation email itself -- this call only starts that process.
func (c *Client) CreateSenderSignature(ctx context.Context, fromEmail, name string) (SenderSignature, error) {
	body := map[string]string{"FromEmail": fromEmail, "Name": name}
	return doJSON[SenderSignature](ctx, c, http.MethodPost, "/senders", body)
}

// GetSenderSignature fetches a sender signature's current state -- used to
// notice Confirmed flipping to true once the customer clicks the link
// Postmark emailed them.
func (c *Client) GetSenderSignature(ctx context.Context, id int) (SenderSignature, error) {
	return doJSON[SenderSignature](ctx, c, http.MethodGet, "/senders/"+strconv.Itoa(id), nil)
}

// DeleteSenderSignature permanently removes a sender signature --
// irreversible.
func (c *Client) DeleteSenderSignature(ctx context.Context, id int) error {
	_, err := doJSON[struct{}](ctx, c, http.MethodDelete, "/senders/"+strconv.Itoa(id), nil)
	return err
}

// DeleteServer permanently deletes a Postmark server -- irreversible, and
// Postmark rejects it (422) unless the server has already been manually
// deactivated first (postmarkapp.com/developer/api/servers-api#delete-server).
func (c *Client) DeleteServer(ctx context.Context, id int) error {
	_, err := doJSON[struct{}](ctx, c, http.MethodDelete, "/servers/"+strconv.Itoa(id), nil)
	return err
}

// IsNotFound reports whether err is Postmark's response for an ID that
// doesn't exist -- lets a delete be treated as already-done rather than a
// failure when retried against a server that's gone.
func IsNotFound(err error) bool {
	var ae *APIError
	return errors.As(err, &ae) && ae.StatusCode == http.StatusNotFound
}

// APIError is returned for any non-2xx response. Postmark's error body is
// {"ErrorCode": int, "Message": string} for every endpoint this client
// calls (postmarkapp.com/developer/api/overview#error-codes).
type APIError struct {
	StatusCode int
	ErrorCode  int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("postmark API: %d (error code %d) %s", e.StatusCode, e.ErrorCode, e.Message)
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
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Postmark-Account-Token", c.accountToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return zero, fmt.Errorf("postmark API request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return zero, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode >= 300 {
		var e struct {
			ErrorCode int    `json:"ErrorCode"`
			Message   string `json:"Message"`
		}
		_ = json.Unmarshal(respBody, &e)
		if e.Message == "" {
			e.Message = string(respBody)
		}
		return zero, &APIError{StatusCode: resp.StatusCode, ErrorCode: e.ErrorCode, Message: e.Message}
	}

	var out T
	if err := json.Unmarshal(respBody, &out); err != nil {
		return zero, fmt.Errorf("decode response body: %w", err)
	}
	return out, nil
}
