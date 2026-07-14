// Command api serves the REST API described in docs/plan.md: orgs and the
// listmonk-tenant instances each org owns.
//
//	@title			listnun API
//	@description	Orgs and the listmonk-tenant instances each org owns. See docs/plan.md for architecture.
//	@version		1.0
//	@BasePath		/api
//
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Bearer token verified against Zitadel's JWKS -- no signup/login endpoints of our own, see docs/plan.md's Auth section.
package main

import (
	"context"
	"log"
	"net/http"

	"listnun/internal/authn"
	"listnun/internal/config"
	"listnun/internal/cryptoutil"
	"listnun/internal/httpapi"
	"listnun/internal/operatorclient"
	"listnun/internal/postmarkclient"
	"listnun/internal/provisioning"
	"listnun/internal/zitadelmgmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("api: connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("api: ping database: %v", err)
	}

	if cfg.ListmonkOperatorToken == "" {
		log.Fatalf("api: LISTMONK_OPERATOR_TOKEN is required")
	}
	op := operatorclient.New(cfg.ListmonkOperatorBaseURL, cfg.ListmonkOperatorToken)

	// Inviting users is optional: only wire up a Zitadel service account
	// if one is configured. Unlike the operator token above, the API runs
	// fine without it -- InviteMember just returns ErrInvitesNotConfigured.
	var zm *zitadelmgmt.Client
	if cfg.ZitadelServiceAccountKeyPath != "" {
		var zmOpts []zitadelmgmt.Option
		if cfg.ZitadelInsecurePort != 0 {
			zmOpts = append(zmOpts, zitadelmgmt.WithInsecurePort(cfg.ZitadelInsecurePort))
		}
		zm, err = zitadelmgmt.New(ctx, cfg.ZitadelDomain, cfg.ZitadelServiceAccountKeyPath, cfg.ZitadelOrgID, zmOpts...)
		if err != nil {
			log.Fatalf("api: connect Zitadel service account: %v", err)
		}
		defer zm.Close()
	}

	// Creating a Postmark server per tenant is optional too: only wire it
	// up if an account token is configured, same fail-open pattern as the
	// Zitadel service account above. Unlike that one, a misconfigured
	// encryption key here is fatal rather than silently skipped -- an
	// operator who set the token clearly intends this step to run, and a
	// bad key would otherwise fail invisibly on every single instance
	// creation instead of at startup.
	var pm *provisioning.PostmarkConfig
	if cfg.PostmarkAccountToken != "" {
		encKey, err := cryptoutil.ParseKey(cfg.PostmarkTokenEncryptionKey)
		if err != nil {
			log.Fatalf("api: parse POSTMARK_TOKEN_ENCRYPTION_KEY: %v", err)
		}
		pm = &provisioning.PostmarkConfig{
			Client:           postmarkclient.New(cfg.PostmarkAccountToken),
			EncryptionKey:    encKey,
			SharedDomainRoot: cfg.PostmarkSharedDomainRoot,
		}
	}

	svc := provisioning.New(pool, op, zm, pm)

	if cfg.ZitadelIssuer == "" {
		log.Fatalf("api: ZITADEL_ISSUER is required")
	}
	verifier, err := authn.NewVerifier(ctx, cfg.ZitadelIssuer, cfg.ZitadelClientID)
	if err != nil {
		log.Fatalf("api: connect to Zitadel issuer: %v", err)
	}

	handler := httpapi.New(svc, verifier)

	log.Printf("api: listening on %s", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, handler); err != nil {
		log.Fatalf("api: serve: %v", err)
	}
}
