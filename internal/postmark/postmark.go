// Package postmark provides a minimal client for the subset of
// Postmark's Account-level API needed to auto-provision a dedicated
// Postmark server (and its SMTP credentials) per tenant. Not a general
// Postmark SDK - only what cmd/operator.go's CreateTenant needs.
package postmark

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// apiURL is Postmark's Account API endpoint for server management. Account
// API calls (as opposed to Server API calls) are authenticated with an
// account-level token and can create new servers - that's why this needs
// a distinct, higher-privilege token from any individual server's own API
// token. A var, not a const, so tests can point it at a local mock server.
var apiURL = "https://api.postmarkapp.com/servers"

// SMTPHost and SMTPPort are Postmark's fixed SMTP relay endpoint. Every
// server on a Postmark account shares this same host - what's
// server-specific is the API token, used as both the SMTP username and
// password.
const (
	SMTPHost = "smtp.postmarkapp.com"
	SMTPPort = 587
)

// Server is the subset of Postmark's server object this package needs.
// See https://postmarkapp.com/developer/api/server-api#create-server.
type Server struct {
	ID        int      `json:"ID"`
	Name      string   `json:"Name"`
	ApiTokens []string `json:"ApiTokens"`
}

// apiError mirrors Postmark's error response shape
// (https://postmarkapp.com/developer/api/overview#error-codes).
type apiError struct {
	ErrorCode int    `json:"ErrorCode"`
	Message   string `json:"Message"`
}

// CreateServer provisions a new Postmark server named name under the
// account identified by accountToken. The returned Server's ApiTokens[0]
// is used as both the SMTP username and password for Postmark's SMTP
// relay - Postmark has no separate "SMTP credentials" concept, the
// server's own API token doubles as both.
func CreateServer(accountToken, name string) (Server, error) {
	body, err := json.Marshal(map[string]string{"Name": name})
	if err != nil {
		return Server{}, err
	}

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return Server{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Postmark-Account-Token", accountToken)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return Server{}, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return Server{}, err
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr apiError
		_ = json.Unmarshal(b, &apiErr)
		return Server{}, fmt.Errorf("postmark API error (HTTP %d, code %d): %s", resp.StatusCode, apiErr.ErrorCode, apiErr.Message)
	}

	var out Server
	if err := json.Unmarshal(b, &out); err != nil {
		return Server{}, fmt.Errorf("error parsing postmark response: %w", err)
	}
	if len(out.ApiTokens) == 0 {
		return Server{}, fmt.Errorf("postmark server %q created but no API token was returned", name)
	}

	return out, nil
}
