// Command worker runs the River-backed provisioning job queue described in
// docs/plan.md.
//
// Bootstrap stub: wires up config and the sqlc-generated database layer so
// the module builds end to end. River client setup and job workers
// (create_postmark_server, provision_k8s_resources, ...) land in the next
// milestone.
package main

import (
	"context"
	"log"

	"listnun/internal/config"
	"listnun/internal/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("worker: connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("worker: ping database: %v", err)
	}

	queries := db.New(pool)
	_ = queries // wired up for use once River jobs land

	log.Printf("worker: connected to database (River wiring not yet implemented)")
}
