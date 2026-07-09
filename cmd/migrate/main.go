// Command migrate applies db/migrations against DATABASE_URL using
// golang-migrate -- see docs/plan.md's Tech stack table. Run from the
// repo root (the source path below is relative), e.g.:
//
//	go run ./cmd/migrate                 # up
//	go run ./cmd/migrate -direction down # down one step
package main

import (
	"errors"
	"flag"
	"log"

	"listnun/internal/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	direction := flag.String("direction", "up", `"up" or "down" (one step)`)
	flag.Parse()

	cfg := config.Load()

	m, err := migrate.New("file://db/migrations", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("migrate: open: %v", err)
	}
	defer m.Close()

	switch *direction {
	case "up":
		err = m.Up()
	case "down":
		err = m.Steps(-1)
	default:
		log.Fatalf("migrate: unknown -direction %q (want up or down)", *direction)
	}
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("migrate: %s: %v", *direction, err)
	}

	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		log.Fatalf("migrate: read version: %v", err)
	}
	log.Printf("migrate: %s complete (version=%d dirty=%v)", *direction, version, dirty)
}
