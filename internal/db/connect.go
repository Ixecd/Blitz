package db

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func NewDB() (*sql.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://blitz:blitz@localhost:5432/blitz?sslmode=disable&search_path=public"
	}

	database, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := database.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	if err := runMigrations(database); err != nil {
		return nil, fmt.Errorf("migrate db: %w", err)
	}

	slog.Info("数据库已连接")
	return database, nil
}

func runMigrations(database *sql.DB) error {
	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("migrate source: %w", err)
	}

	driver, err := postgres.WithInstance(database, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("migrate driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "postgres", driver)
	if err != nil {
		return fmt.Errorf("migrate init: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}

	v, _, _ := m.Version()
	slog.Info("数据库迁移完成", "version", v)
	return nil
}
