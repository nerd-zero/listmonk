// Package authn verifies the bearer token the frontend sends after
// authenticating directly against Zitadel (OIDC Authorization Code flow,
// per docs/plan.md) -- no signup/login/refresh endpoints of our own.
package authn

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
)

// Claims is what the rest of the app needs from a verified token, ready to
// pass straight into provisioning.Service.JITProvisionUser.
type Claims struct {
	Subject     string
	Email       string
	DisplayName string
}

// Verifier checks a bearer token against Zitadel's JWKS.
type Verifier struct {
	verifier *oidc.IDTokenVerifier
}

// NewVerifier fetches the issuer's OIDC discovery document (JWKS location,
// etc.) once at startup. Fails fast if the issuer is unreachable or
// misconfigured -- there's no safe way to run the API without this.
func NewVerifier(ctx context.Context, issuer, clientID string) (*Verifier, error) {
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, err
	}
	return &Verifier{verifier: provider.Verifier(&oidc.Config{ClientID: clientID})}, nil
}

func (v *Verifier) Verify(ctx context.Context, bearerToken string) (Claims, error) {
	idToken, err := v.verifier.Verify(ctx, bearerToken)
	if err != nil {
		return Claims{}, err
	}

	var raw struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := idToken.Claims(&raw); err != nil {
		return Claims{}, err
	}

	return Claims{Subject: idToken.Subject, Email: raw.Email, DisplayName: raw.Name}, nil
}

// BearerToken extracts the token from a request's Authorization header.
func BearerToken(r *http.Request) (string, error) {
	h := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(h, prefix) {
		return "", errors.New("missing bearer token")
	}
	return strings.TrimPrefix(h, prefix), nil
}
