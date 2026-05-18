// Package store — слой данных (аналог SQLAlchemy repository / Prisma client).
package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite" // регистрирует драйвер "sqlite"
)

// Store инкапсулирует подключение к SQLite.
type Store struct {
	db *sql.DB
}

// Open открывает БД и применяет миграции схемы.
func Open(dbPath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o700); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	db.SetMaxOpenConns(1) // SQLite — один писатель за раз

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) migrate() error {
	schema := `
CREATE TABLE IF NOT EXISTS server_settings (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  listen_port INTEGER NOT NULL DEFAULT 51821,
  subnet_cidr TEXT NOT NULL DEFAULT '10.99.0.0/24',
  server_private_key TEXT NOT NULL,
  server_public_key TEXT NOT NULL,
  endpoint_host TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE,
  client_private_key TEXT NOT NULL,
  client_public_key TEXT NOT NULL,
  assigned_ip TEXT NOT NULL UNIQUE,
  active INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL,
  revoked_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_users_active ON users(active);
`
	if _, err := s.db.Exec(schema); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}

// Close закрывает соединение.
func (s *Store) Close() error {
	return s.db.Close()
}

// ServerSettings — конфигурация VPN-сервера (одна запись).
type ServerSettings struct {
	ListenPort       int
	SubnetCIDR       string
	ServerPrivateKey string
	ServerPublicKey  string
	EndpointHost     string
	CreatedAt        time.Time
}

// User — VPN-пользователь (peer).
type User struct {
	ID               int64
	Name             string
	ClientPrivateKey string
	ClientPublicKey  string
	AssignedIP       string
	Active           bool
	CreatedAt        time.Time
	RevokedAt        *time.Time
}
