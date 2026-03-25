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

	if err := initAdmin(database); err != nil {
		slog.Warn("初始化管理员失败", "err", err)
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

// initAdmin 通过 ADMIN_EMAIL 环境变量注入初始管理员
// 幂等：重复启动不会重复授权，用户不存在时静默跳过
func initAdmin(db *sql.DB) error {
	email := os.Getenv("ADMIN_EMAIL")
	if email == "" {
		return nil
	}

	var userID int64
	err := db.QueryRow(`SELECT id FROM users WHERE email = $1`, email).Scan(&userID)
	if err == sql.ErrNoRows {
		slog.Debug("ADMIN_EMAIL 用户不存在，跳过授权", "email", email)
		return nil
	}
	if err != nil {
		return fmt.Errorf("查询用户失败: %w", err)
	}

	var roleID int64
	err = db.QueryRow(`SELECT id FROM roles WHERE name = 'admin'`).Scan(&roleID)
	if err != nil {
		return fmt.Errorf("查询 admin 角色失败: %w", err)
	}

	_, err = db.Exec(`
		INSERT INTO user_roles (user_id, role_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, userID, roleID)
	if err != nil {
		return fmt.Errorf("授权失败: %w", err)
	}

	slog.Info("已授予管理员权限", "email", email, "user_id", userID)
	return nil
}
