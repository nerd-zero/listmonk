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
// through every Core method to actually call WithTenant is issue #40's
// job, not this one.
//
// opts is passed straight to BeginTxx (nil means BeginTxx's own default,
// same as calling it directly) - callers that need e.g. ReadOnly: true as
// a security control (see subscribers.go's arbitrary-query features) can
// still get it instead of losing that guarantee to this helper.
func (c *Core) WithTenant(ctx context.Context, tenantID int, opts *sql.TxOptions, fn func(tx *sqlx.Tx) error) error {
	tx, err := c.db.BeginTxx(ctx, opts)
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

// stmtx rebinds a pool-level prepared statement (c.q.X) to run within tx,
// so it participates in WithTenant's transaction (and therefore its
// app.current_tenant setting) instead of picking a random connection from
// the pool the way calling the statement directly would.
//
// This must be used instead of tx.Stmtx(stmt) directly: cmd/init.go opens
// the pool with .Unsafe() (several models don't map every returned column,
// e.g. campaigns.go's own "Unsafe to ignore scanning fields not present in
// models.Campaigns" - and every table has an unmapped tenant_id column
// since phase 1 unless its model struct happens to declare one). BeginTxx
// correctly copies that Unsafe flag onto the resulting *sqlx.Tx, but
// sqlx.Tx.Stmtx does NOT copy it onto the *sqlx.Stmt it derives - found by
// hitting "missing destination name tenant_id" errors that only appeared
// on the tx-rebound path, not on the original pool-level statement or on
// tx.Select/tx.Get called directly. Confirmed via sqlx v1.4.0 source
// (Tx.Stmtx: `&Stmt{Stmt: tx.Stmt(s), Mapper: tx.Mapper}`, no unsafe field
// set) rather than assumed.
func stmtx(tx *sqlx.Tx, stmt *sqlx.Stmt) *sqlx.Stmt {
	return tx.Stmtx(stmt).Unsafe()
}
