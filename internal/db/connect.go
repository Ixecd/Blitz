package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func NewDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	if err := initSchema(db); err != nil {
		return nil, err
	}
	return db, nil
}

func initSchema(db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS deposit_addresses (
            id         INTEGER PRIMARY KEY AUTOINCREMENT,
            user_id    TEXT NOT NULL,
            address    TEXT NOT NULL UNIQUE,
            chain      TEXT NOT NULL,
            path       TEXT NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )`,
		`CREATE TABLE IF NOT EXISTS deposits (
            id         INTEGER PRIMARY KEY AUTOINCREMENT,
            tx_id      TEXT NOT NULL UNIQUE,
            address    TEXT NOT NULL,
            user_id    TEXT NOT NULL,
            amount     REAL NOT NULL,
            height     INTEGER NOT NULL,
            confirmed  INTEGER NOT NULL DEFAULT 0,
            chain      TEXT NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )`,
		`CREATE TABLE IF NOT EXISTS withdrawals (
            id         INTEGER PRIMARY KEY AUTOINCREMENT,
            tx_id      TEXT,
            address    TEXT NOT NULL,
            user_id    TEXT NOT NULL,
            amount     REAL NOT NULL,
            fee        REAL NOT NULL DEFAULT 0,
            status     TEXT NOT NULL DEFAULT 'pending',
            chain      TEXT NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )`,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}
