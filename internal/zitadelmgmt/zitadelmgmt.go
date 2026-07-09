// Package zitadelmgmt wraps the official Zitadel Go SDK's machine-user
// ("service account") authentication and User Service v2 API, used to
// invite a new person into an org directly from listnun's own UI rather
// than relying on Zitadel's own self-registration screen.
//
// Why this exists at all: internal/authn only ever sees bearer tokens from
// people who have already signed themselves into Zitadel -- the backend
// has no human session to act through. Inviting someone who doesn't have
// an account yet means calling Zitadel's Management-plane API as its own
// identity, which is exactly what Zitadel calls a "service user" (a
// machine account authenticated by a private-key JWT, not a password or
// shared secret). This package is that identity's client.
//
// Live-verified (2026-07-09) against a real Zitadel v4.12.1 instance
// (docker-compose.yml's dev stack): a machine key generated via the real
// Management API authenticates through DefaultServiceUserAuthentication,
// and InviteHuman below successfully creates a real human user. See
// docs/plan.md's Implementation log for the exact steps.
package zitadelmgmt

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/zitadel/oidc/v3/pkg/oidc"
	"github.com/zitadel/zitadel-go/v3/pkg/client"
	objectv2 "github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/object/v2"
	userv2 "github.com/zitadel/zitadel-go/v3/pkg/client/zitadel/user/v2"
	"github.com/zitadel/zitadel-go/v3/pkg/zitadel"
)

// Client is listnun's own identity in Zitadel, not any end user's.
type Client struct {
	zc    *client.Client
	orgID string
}

type options struct {
	insecurePort uint16 // 0 means TLS (production default)
}

type Option func(*options)

// WithInsecurePort connects over plain HTTP instead of TLS, on the given
// port -- for a local docker-compose Zitadel run with --tlsMode disabled.
// Never use in production (same warning Zitadel's own SDK option carries).
func WithInsecurePort(port uint16) Option {
	return func(o *options) { o.insecurePort = port }
}

// New authenticates as the service account described by the key file at
// keyPath -- the same key.json Zitadel's console (or Management API --
// see docs/plan.md's Implementation log) hands you when you create a
// machine user with a key. domain is the Zitadel instance host with no
// scheme (e.g. "your-instance.zitadel.cloud", or "localhost" for dev with
// WithInsecurePort). orgID is the Zitadel Organization new users are
// created under; listnun uses a single shared Zitadel org for every
// customer's people, distinct from listnun's own per-customer orgs table
// (see docs/plan.md).
func New(ctx context.Context, domain, keyPath, orgID string, opts ...Option) (*Client, error) {
	if orgID == "" {
		return nil, errors.New("zitadelmgmt: orgID is required")
	}

	var o options
	for _, opt := range opts {
		opt(&o)
	}

	var zOpts []zitadel.Option
	if o.insecurePort != 0 {
		zOpts = append(zOpts, zitadel.WithInsecure(strconv.Itoa(int(o.insecurePort))))
	}
	z := zitadel.New(domain, zOpts...)
	auth := client.DefaultServiceUserAuthentication(keyPath, oidc.ScopeOpenID, client.ScopeZitadelAPI())

	zc, err := client.New(ctx, z, client.WithAuth(auth))
	if err != nil {
		return nil, fmt.Errorf("zitadelmgmt: connect: %w", err)
	}
	return &Client{zc: zc, orgID: orgID}, nil
}

func (c *Client) Close() error {
	return c.zc.Close()
}

// InviteHuman creates a new human user in Zitadel and returns its user
// ID -- this becomes the new users.zitadel_subject row the caller inserts
// locally alongside an org_members row, so the person's *next* login (once
// they've completed Zitadel's own verification/password-set flow) is just
// an ordinary JITProvisionUser lookup that finds the row already there,
// no special-casing required.
//
// Verification is left unset, so Zitadel sends its own default
// verification email. Swapping this for SetHumanEmail_ReturnCode to
// deliver a listnun-branded invite link instead (the same pattern already
// used for listmonk's setup_url) needs listnun's own outbound email, which
// isn't built yet -- see docs/plan.md's Postmark section.
func (c *Client) InviteHuman(ctx context.Context, email, displayName string) (string, error) {
	given, family := splitName(displayName, email)

	resp, err := c.zc.UserServiceV2().AddHumanUser(ctx, &userv2.AddHumanUserRequest{
		Organization: &objectv2.Organization{Org: &objectv2.Organization_OrgId{OrgId: c.orgID}},
		Profile: &userv2.SetHumanProfile{
			GivenName:   given,
			FamilyName:  family,
			DisplayName: &displayName,
		},
		Email: &userv2.SetHumanEmail{Email: email},
	})
	if err != nil {
		return "", fmt.Errorf("zitadelmgmt: add human user: %w", err)
	}
	return resp.UserId, nil
}

// splitName does the best it can with a single display-name string --
// Zitadel's profile requires given/family name separately, but listnun
// only ever collects one name field. Falls back to the email's local part
// if displayName is empty, since both fields are required by Zitadel.
func splitName(displayName, email string) (given, family string) {
	name := strings.TrimSpace(displayName)
	if name == "" {
		if at := strings.IndexByte(email, '@'); at > 0 {
			name = email[:at]
		} else {
			name = email
		}
	}
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return name, name
	}
	if len(parts) == 1 {
		return parts[0], parts[0]
	}
	return parts[0], strings.Join(parts[1:], " ")
}
