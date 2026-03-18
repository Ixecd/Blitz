package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewDB(path string) (*sql.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	driver := "pgx"

	if dsn == "" {
		// 本地开发默认值
		dsn = "postgres://blitz:blitz@localhost:5432/blitz?sslmode=disable"
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return db, nil
}
