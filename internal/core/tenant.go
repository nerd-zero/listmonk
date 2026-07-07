package core

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// WithTenant runs fn inside a transaction with `app.current_tenant` set for
// that transaction's duration (via set_config's third "is_local" argument,
// Postgres's equivalent of `SET LOCAL`), so the setting is automatically
// cleared when the transaction ends rather than leaking onto a
// subsequently reused connection.
//
// This is safe under Go's own connection pooling: database/sql binds a
// transaction to a single physical connection for its entire lifetime and
// only returns that connection to the pool after Commit/Rollback, by which
// point Postgres has already reset the transaction-local setting. See
// TestWithTenant_ConcurrentIsolation for a concurrency spike proving this
// holds for this specific driver/pool combination.
//
// This does NOT hold if listmonk is ever deployed behind an external
// transaction-pooling proxy (e.g. PgBouncer in transaction-pooling mode)
// that can multiplex statements from one logical transaction across
// different physical connections - that combination is out of scope here
// and must be verified separately before relying on it.
//
// The RLS policies from phase 2 (v6.5.0) are permissive while
// `app.current_tenant` is unset, so this helper is additive: existing
// Core methods that don't yet call it are unaffected. Threading tenantID
// through every Core method to actually call WithTenant is phase 4's job,
// not this one.
func (c *Core) WithTenant(ctx context.Context, tenantID int, fn func(tx *sqlx.Tx) error) error {
	tx, err := c.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `SELECT set_config('app.current_tenant', $1::TEXT, true)`, tenantID); err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}
