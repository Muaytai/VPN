package store

import "fmt"

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
		return fmt.Errorf("migrate schema: %w", err)
	}
	return s.migrateV2()
}

func (s *Store) migrateV2() error {
	columns := map[string]string{
		"full_tunnel":   "INTEGER NOT NULL DEFAULT 1",
		"wan_interface": "TEXT NOT NULL DEFAULT ''",
		"default_dns":   "TEXT NOT NULL DEFAULT '1.1.1.1, 8.8.8.8'",
	}
	existing, err := s.tableColumns("server_settings")
	if err != nil {
		return err
	}
	for name, ddl := range columns {
		if existing[name] {
			continue
		}
		q := fmt.Sprintf("ALTER TABLE server_settings ADD COLUMN %s %s", name, ddl)
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("add column %s: %w", name, err)
		}
	}
	return nil
}

func (s *Store) tableColumns(table string) (map[string]bool, error) {
	rows, err := s.db.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols := map[string]bool{}
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue any
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			return nil, err
		}
		cols[name] = true
	}
	return cols, rows.Err()
}
