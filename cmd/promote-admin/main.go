// Command promote-admin grants or revokes is_super_admin for a user by
// email -- the only way to do so, by deliberate design (see that column's
// doc comment in db/migrations/0001_init.up.sql: no self-service grant).
// Run from the repo root, e.g.:
//
//	go run ./cmd/promote-admin -email you@example.com
//	go run ./cmd/promote-admin -email you@example.com -revoke
package main

import (
	"context"
	"errors"
	"flag"
	"log"

	"listnun/internal/config"
	"listnun/internal/db"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	email := flag.String("email", "", "email of the user to promote (required)")
	revoke := flag.Bool("revoke", false, "revoke super admin instead of granting it")
	flag.Parse()

	if *email == "" {
		log.Fatal("promote-admin: -email is required")
	}

	cfg := config.Load()

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("promote-admin: connect to database: %v", err)
	}
	defer pool.Close()

	q := db.New(pool)
	user, err := q.SetUserSuperAdmin(ctx, db.SetUserSuperAdminParams{Email: *email, IsSuperAdmin: !*revoke})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Fatalf("promote-admin: no user found with email %q", *email)
		}
		log.Fatalf("promote-admin: %v", err)
	}

	verb := "granted"
	if *revoke {
		verb = "revoked"
	}
	log.Printf("promote-admin: super admin %s for %s (user id %s)", verb, user.Email, user.ID)
}
