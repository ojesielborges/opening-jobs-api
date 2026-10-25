package postgres

import (
	"context"
	"fmt"

	"openingjobs/internal/tenant"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type TenantDB struct{ db *sqlx.DB }

func NewTenantDB(db *sqlx.DB) TenantDB { return TenantDB{db: db} }

func (db TenantDB) WithinTransaction(ctx context.Context, operation func(*sqlx.Tx) error) error {
	current, err := tenant.FromContext(ctx)
	if err != nil {
		return err
	}

	tx, err := db.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tenant transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "SET LOCAL search_path TO "+pq.QuoteIdentifier(current.SchemaName)+", public"); err != nil {
		return fmt.Errorf("set tenant schema: %w", err)
	}
	if err := operation(tx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tenant transaction: %w", err)
	}
	return nil
}
