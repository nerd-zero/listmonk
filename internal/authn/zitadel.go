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
	"golang.org/x/oauth2"
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
	provider *oidc.Provider
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
	return &Verifier{provider: provider, verifier: provider.Verifier(&oidc.Config{ClientID: clientID})}, nil
}

// Verify checks bearerToken -- the frontend's OAuth2 access token, not an ID
// token (see web/src/api/mutator.ts) -- against Zitadel's JWKS, then calls
// the userinfo endpoint for email/name: Zitadel's JWT access tokens carry
// only sub/aud/scope, never profile claims, regardless of the OIDC client's
// "userinfo inside ID token" setting -- that only affects ID tokens, which
// never leave the frontend. The userinfo endpoint is the only place a
// resource server can get profile data for a bearer access token.
func (v *Verifier) Verify(ctx context.Context, bearerToken string) (Claims, error) {
	idToken, err := v.verifier.Verify(ctx, bearerToken)
	if err != nil {
		return Claims{}, err
	}

	userInfo, err := v.provider.UserInfo(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: bearerToken}))
	if err != nil {
		return Claims{}, err
	}
	var raw struct {
		Name string `json:"name"`
	}
	if err := userInfo.Claims(&raw); err != nil {
		return Claims{}, err
	}

	return Claims{Subject: idToken.Subject, Email: userInfo.Email, DisplayName: raw.Name}, nil
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
