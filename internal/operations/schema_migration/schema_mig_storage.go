package schemamigration

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type SchemaMigrationStorage struct {
	db         *sqlx.DB
	schemaName string
}

func NewSchemaMigrationStorage(db *sqlx.DB, schemaName string) *SchemaMigrationStorage {
	return &SchemaMigrationStorage{db: db, schemaName: schemaName}
}

func (storage *SchemaMigrationStorage) appliedMigrations() (map[string]bool, error) {
	applied := make(map[string]bool)
	err := storage.withTransaction(func(tx *sqlx.Tx) error {
		rows, err := tx.Query(`SELECT filename FROM schema_migrations`)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var filename string
			if err := rows.Scan(&filename); err != nil {
				return err
			}
			applied[filename] = true
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return applied, nil
}

func (storage *SchemaMigrationStorage) markMigrationApplied(filename string) error {
	return storage.withTransaction(func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`INSERT INTO schema_migrations (filename) VALUES ($1)`, filename)
		return err
	})
}

func (storage *SchemaMigrationStorage) unmarkMigration(filename string) error {
	return storage.withTransaction(func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`DELETE FROM schema_migrations WHERE filename = $1`, filename)
		return err
	})
}

func (storage *SchemaMigrationStorage) execSql(sql string) error {
	return storage.withTransaction(func(tx *sqlx.Tx) error {
		_, err := tx.Exec(sql)
		return err
	})
}

func (storage *SchemaMigrationStorage) withTransaction(operation func(*sqlx.Tx) error) error {
	tx, err := storage.db.Beginx()
	if err != nil {
		return fmt.Errorf("begin migration transaction: %w", err)
	}
	defer tx.Rollback()
	if storage.schemaName != "" {
		if _, err := tx.Exec("SET LOCAL search_path TO " + pq.QuoteIdentifier(storage.schemaName) + ", public"); err != nil {
			return fmt.Errorf("set migration schema: %w", err)
		}
	}
	if err := operation(tx); err != nil {
		return err
	}
	return tx.Commit()
}
