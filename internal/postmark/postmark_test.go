package postmark

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Postmark-Account-Token"); got != "test-account-token" {
			t.Errorf("unexpected account token header: %q", got)
		}

		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("error decoding request body: %v", err)
		}
		if body["Name"] != "acme" {
			t.Errorf("unexpected server name: %q", body["Name"])
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Server{ID: 42, Name: "acme", ApiTokens: []string{"fake-server-token"}})
	}))
	defer srv.Close()

	origURL := apiURL
	apiURL = srv.URL
	defer func() { apiURL = origURL }()

	out, err := CreateServer("test-account-token", "acme")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.ID != 42 || out.Name != "acme" || len(out.ApiTokens) != 1 || out.ApiTokens[0] != "fake-server-token" {
		t.Errorf("unexpected server: %+v", out)
	}
}

func TestCreateServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(apiError{ErrorCode: 10, Message: "invalid API token"})
	}))
	defer srv.Close()

	origURL := apiURL
	apiURL = srv.URL
	defer func() { apiURL = origURL }()

	_, err := CreateServer("bad-token", "acme")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}
