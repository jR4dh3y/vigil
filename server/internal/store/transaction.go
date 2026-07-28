package store

import (
	"context"
	"database/sql"
	"fmt"
)

type transactionBeginner interface {
	BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
}

// InTransaction runs fn with transaction-bound queries and commits only when
// fn succeeds.
func InTransaction(ctx context.Context, q *Queries, fn func(*Queries) error) error {
	if q == nil {
		return fmt.Errorf("queries are not configured")
	}
	db, ok := q.db.(transactionBeginner)
	if !ok {
		return fmt.Errorf("queries do not support transactions")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := fn(q.WithTx(tx)); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
