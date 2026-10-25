package tenant

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Provisioner struct{ db *sqlx.DB }

func NewProvisioner(db *sqlx.DB) *Provisioner { return &Provisioner{db: db} }

func (p *Provisioner) Provision(ctx context.Context, slug, apiKey, migrationsPath string) error {
	if len(apiKey) < 32 {
		return fmt.Errorf("tenant API key must contain at least 32 characters")
	}
	schemaName, err := SchemaName(slug)
	if err != nil {
		return err
	}
	files, err := filepath.Glob(filepath.Join(migrationsPath, "*_up.sql"))
	if err != nil {
		return fmt.Errorf("list tenant migrations: %w", err)
	}
	sort.Strings(files)
	if len(files) == 0 {
		return fmt.Errorf("no tenant migrations found in %q", migrationsPath)
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

	var applied []string
	if err := tx.SelectContext(ctx, &applied, `SELECT filename FROM schema_migrations`); err != nil {
		return fmt.Errorf("read tenant migrations: %w", err)
	}
	done := make(map[string]struct{}, len(applied))
	for _, name := range applied {
		done[name] = struct{}{}
	}
	for _, file := range files {
		name := filepath.Base(file)
		if _, ok := done[name]; ok {
			continue
		}
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx, strings.TrimSpace(string(content))); err != nil {
			return fmt.Errorf("apply tenant migration %s: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (filename) VALUES ($1)`, name); err != nil {
			return fmt.Errorf("record tenant migration %s: %w", name, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tenant provisioning: %w", err)
	}
	return nil
}
