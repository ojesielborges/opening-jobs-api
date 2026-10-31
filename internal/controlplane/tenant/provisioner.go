package tenant

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Provisioner struct{ db *sqlx.DB }

func NewProvisioner(db *sqlx.DB) *Provisioner { return &Provisioner{db: db} }

func (p *Provisioner) Provision(ctx context.Context, slug, apiKey string) error {
	if len(apiKey) < 32 {
		return fmt.Errorf("tenant API key must contain at least 32 characters")
	}
	schemaName, err := SchemaName(slug)
	if err != nil {
		return err
	}
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tenant provisioning: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `CREATE SCHEMA IF NOT EXISTS control`); err != nil {
		return fmt.Errorf("create control schema: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`); err != nil {
		return fmt.Errorf("enable UUID extension: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS control.tenants (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), slug VARCHAR(50) UNIQUE NOT NULL,
		schema_name VARCHAR(63) UNIQUE NOT NULL, api_key_hash CHAR(64) NOT NULL,
		active BOOLEAN NOT NULL DEFAULT TRUE, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		CHECK (slug ~ '^[a-z0-9_]{2,50}$'),
		CHECK (schema_name ~ '^tenant_[a-z0-9_]{2,50}$'))`); err != nil {
		return fmt.Errorf("create tenant registry: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "CREATE SCHEMA IF NOT EXISTS "+pq.QuoteIdentifier(schemaName)); err != nil {
		return fmt.Errorf("create tenant schema: %w", err)
	}

	sum := sha256.Sum256([]byte(apiKey))
	if _, err := tx.ExecContext(ctx, `INSERT INTO control.tenants (slug, schema_name, api_key_hash)
		VALUES ($1, $2, $3)
		ON CONFLICT (slug) DO UPDATE SET schema_name = EXCLUDED.schema_name, api_key_hash = EXCLUDED.api_key_hash, active = TRUE`,
		slug, schemaName, hex.EncodeToString(sum[:])); err != nil {
		return fmt.Errorf("register tenant: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "SET LOCAL search_path TO "+pq.QuoteIdentifier(schemaName)+", public"); err != nil {
		return fmt.Errorf("set tenant schema: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		filename VARCHAR(255) PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`); err != nil {
		return fmt.Errorf("create tenant migration registry: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tenant provisioning: %w", err)
	}
	return nil
}
