package store

import (
	"database/sql"
	"fmt"
	"time"
)

// HasServerSettings проверяет, выполнялась ли команда init.
func (s *Store) HasServerSettings() (bool, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM server_settings`).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("count server_settings: %w", err)
	}
	return count > 0, nil
}

// SaveServerSettings сохраняет настройки сервера (вызывается один раз при init).
func (s *Store) SaveServerSettings(settings ServerSettings) error {
	_, err := s.db.Exec(`
INSERT INTO server_settings (
  id, listen_port, subnet_cidr, server_private_key, server_public_key,
  endpoint_host, created_at, full_tunnel, wan_interface, default_dns
) VALUES (1, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`,
		settings.ListenPort,
		settings.SubnetCIDR,
		settings.ServerPrivateKey,
		settings.ServerPublicKey,
		settings.EndpointHost,
		settings.CreatedAt.UTC().Format(time.RFC3339),
		boolToInt(settings.FullTunnel),
		settings.WANInterface,
		settings.DefaultDNS,
	)
	if err != nil {
		return fmt.Errorf("insert server_settings: %w", err)
	}
	return nil
}

// UpdateServerNetwork обновляет режим full-tunnel и WAN-интерфейс.
func (s *Store) UpdateServerNetwork(fullTunnel bool, wanInterface, defaultDNS string) error {
	_, err := s.db.Exec(`
UPDATE server_settings SET full_tunnel = ?, wan_interface = ?, default_dns = ? WHERE id = 1
`, boolToInt(fullTunnel), wanInterface, defaultDNS)
	if err != nil {
		return fmt.Errorf("update server network: %w", err)
	}
	return nil
}

// GetServerSettings возвращает настройки сервера.
func (s *Store) GetServerSettings() (ServerSettings, error) {
	var (
		settings     ServerSettings
		createdAt    string
		fullTunnel   int
		wanInterface sql.NullString
		defaultDNS   sql.NullString
	)
	err := s.db.QueryRow(`
SELECT listen_port, subnet_cidr, server_private_key, server_public_key, endpoint_host, created_at,
       COALESCE(full_tunnel, 1), COALESCE(wan_interface, ''), COALESCE(default_dns, '1.1.1.1, 8.8.8.8')
FROM server_settings WHERE id = 1
`).Scan(
		&settings.ListenPort,
		&settings.SubnetCIDR,
		&settings.ServerPrivateKey,
		&settings.ServerPublicKey,
		&settings.EndpointHost,
		&createdAt,
		&fullTunnel,
		&wanInterface,
		&defaultDNS,
	)
	if err == sql.ErrNoRows {
		return ServerSettings{}, fmt.Errorf("server not initialized: run `vpnctl init`")
	}
	if err != nil {
		return ServerSettings{}, fmt.Errorf("select server_settings: %w", err)
	}

	settings.FullTunnel = fullTunnel == 1
	settings.WANInterface = wanInterface.String
	settings.DefaultDNS = defaultDNS.String
	if settings.DefaultDNS == "" {
		settings.DefaultDNS = "1.1.1.1, 8.8.8.8"
	}

	t, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return ServerSettings{}, fmt.Errorf("parse created_at: %w", err)
	}
	settings.CreatedAt = t
	return settings, nil
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
