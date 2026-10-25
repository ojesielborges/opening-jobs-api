package config

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func InitializeDB() (*sqlx.DB, error) {
	var err error
	var resultDB *sqlx.DB

	urlDB := GetEnv("URL_DB", "")
	if urlDB == "" {
		return nil, fmt.Errorf("URL_DB is required")
	}

	resultDB, err = sqlx.Open("postgres", urlDB)
	if err != nil {
		err := fmt.Errorf("database connection error: %v", err)
		return nil, err
	}
	resultDB.SetMaxOpenConns(20)
	resultDB.SetMaxIdleConns(10)
	resultDB.SetConnMaxLifetime(30 * time.Minute)
	resultDB.SetConnMaxIdleTime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = resultDB.PingContext(ctx); err != nil {
		resultDB.Close()
		return nil, fmt.Errorf("database ping: %w", err)
	}
	_, err = resultDB.ExecContext(ctx, `CREATE SCHEMA IF NOT EXISTS app; SET search_path TO app, public;`)
	if err != nil {
		resultDB.Close()
		return nil, fmt.Errorf("configure database schema: %w", err)
	}

	return resultDB, nil
}
