package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewDB(path string) (*sql.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://blitz:blitz@localhost:5432/blitz?sslmode=disable"
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	if err := runSeed(db); err != nil {
		return nil, fmt.Errorf("seed db: %w", err)
	}
	return db, nil
}

func runSeed(db *sql.DB) error {
	seed := `
INSERT INTO withdrawal_limits (level, level_name, btc_daily, eth_daily, min_deposit) VALUES
(0, '普通用户',  '2.00000000',   '50.00000000',   '0.00000000'),
(1, '白银用户',  '10.00000000',  '200.00000000',  '1.00000000'),
(2, '黄金用户',  '50.00000000',  '1000.00000000', '10.00000000'),
(3, '钻石用户',  '200.00000000', '5000.00000000', '50.00000000')
ON CONFLICT (level) DO NOTHING;`
	_, err := db.Exec(seed)
	return err
}
