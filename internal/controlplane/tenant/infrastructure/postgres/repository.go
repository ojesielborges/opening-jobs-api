package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"

	"openingjobs/internal/controlplane/tenant"

	"github.com/jmoiron/sqlx"
)

type Repository struct{ db *sqlx.DB }

func NewRepository(db *sqlx.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Resolve(ctx context.Context, slug, apiKey string) (tenant.Tenant, error) {
	if slug == "" || apiKey == "" {
		return tenant.Tenant{}, tenant.ErrUnauthenticated
	}

	sum := sha256.Sum256([]byte(apiKey))
	var result tenant.Tenant
	err := r.db.GetContext(ctx, &result, `
		SELECT id, slug, schema_name
		FROM control.tenants
		WHERE slug = $1 AND api_key_hash = $2 AND active = TRUE`, slug, hex.EncodeToString(sum[:]))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return tenant.Tenant{}, tenant.ErrUnauthenticated
		}
		return tenant.Tenant{}, err
	}
	if !result.Valid() {
		return tenant.Tenant{}, tenant.ErrUnauthenticated
	}
	return result, nil
}
