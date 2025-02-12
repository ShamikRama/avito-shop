package storage

import (
	"avito-shop/internal/config"
	"database/sql"
	"fmt"
	"github.com/pressly/goose"

	_ "github.com/lib/pq"
)

func InitPostgres(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.PgcConnString)
	if err != nil {
		return nil, fmt.Errorf("storage.psql.New: failed to open database connection: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("storage.psql.New: failed to ping database: %w", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return nil, fmt.Errorf("storage.psql.New: failed to set dialect: %w", err)
	}

	if err = goose.Up(db, "./migrations"); err != nil {
		return nil, fmt.Errorf("storage.psql.New: failed to up migrations: %w", err)
	}

	return db, nil
}
