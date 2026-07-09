// Command api serves the REST API described in docs/plan.md: orgs and the
// listmonk-tenant instances each org owns.
package main

import (
	"context"
	"log"
	"net/http"

	"listnun/internal/authn"
	"listnun/internal/config"
	"listnun/internal/httpapi"
	"listnun/internal/operatorclient"
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

	svc := provisioning.New(pool, op, zm)

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
