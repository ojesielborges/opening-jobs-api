package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"openingjobs/internal/opening/domain"
	platformpostgres "openingjobs/internal/platform/postgres"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Repository struct{ tenantDB platformpostgres.TenantDB }

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{tenantDB: platformpostgres.NewTenantDB(db)}
}

const fields = "id, role, company, location, remote, link, salary"

func (r *Repository) Create(ctx context.Context, o domain.Opening) (domain.Opening, error) {
	const query = `INSERT INTO opening (role, company, location, remote, link, salary)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING ` + fields
	var result domain.Opening
	err := r.tenantDB.WithinTransaction(ctx, func(tx *sqlx.Tx) error {
		var err error
		result, err = r.scanOne(ctx, tx, query, o.Role, o.Company, o.Location, o.Remote, o.Link, o.Salary)
		return err
	})
	return result, err
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (domain.Opening, error) {
	var result domain.Opening
	err := r.tenantDB.WithinTransaction(ctx, func(tx *sqlx.Tx) error {
		var err error
		result, err = r.scanOne(ctx, tx, `SELECT `+fields+` FROM opening WHERE id = $1`, id)
		return err
	})
	return result, err
}

func (r *Repository) List(ctx context.Context) ([]domain.Opening, error) {
	var openings []domain.Opening
	err := r.tenantDB.WithinTransaction(ctx, func(tx *sqlx.Tx) error {
		if err := tx.SelectContext(ctx, &openings, `SELECT `+fields+` FROM opening ORDER BY id`); err != nil {
			return fmt.Errorf("list openings: %w", err)
		}
		return nil
	})
	return openings, err
}

func (r *Repository) Update(ctx context.Context, o domain.Opening) (domain.Opening, error) {
	const query = `UPDATE opening SET role=$1, company=$2, location=$3, remote=$4, link=$5, salary=$6
		WHERE id=$7 RETURNING ` + fields
	var result domain.Opening
	err := r.tenantDB.WithinTransaction(ctx, func(tx *sqlx.Tx) error {
		var err error
		result, err = r.scanOne(ctx, tx, query, o.Role, o.Company, o.Location, o.Remote, o.Link, o.Salary, o.ID)
		return err
	})
	return result, err
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.tenantDB.WithinTransaction(ctx, func(tx *sqlx.Tx) error {
		result, err := tx.ExecContext(ctx, `DELETE FROM opening WHERE id = $1`, id)
		if err != nil {
			return fmt.Errorf("delete opening: %w", err)
		}
		count, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("inspect deletion: %w", err)
		}
		if count == 0 {
			return domain.ErrNotFound
		}
		return nil
	})
}

func (r *Repository) scanOne(ctx context.Context, db sqlx.QueryerContext, query string, args ...any) (domain.Opening, error) {
	var opening domain.Opening
	if err := sqlx.GetContext(ctx, db, &opening, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Opening{}, domain.ErrNotFound
		}
		return domain.Opening{}, fmt.Errorf("query opening: %w", err)
	}
	return opening, nil
}
