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
	FullTunnel       bool
	WANInterface     string
	DefaultDNS       string
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
