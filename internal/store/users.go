package store

import (
	"database/sql"
	"fmt"
	"net"
	"time"
)

// CreateUser добавляет пользователя с уже вычисленным IP.
func (s *Store) CreateUser(u User) (User, error) {
	res, err := s.db.Exec(`
INSERT INTO users (name, client_private_key, client_public_key, assigned_ip, active, created_at)
VALUES (?, ?, ?, ?, 1, ?)
`,
		u.Name,
		u.ClientPrivateKey,
		u.ClientPublicKey,
		u.AssignedIP,
		u.CreatedAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return User{}, fmt.Errorf("insert user: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return User{}, fmt.Errorf("last insert id: %w", err)
	}
	u.ID = id
	u.Active = true
	return u, nil
}

// ListUsers возвращает всех пользователей (включая отозванных).
func (s *Store) ListUsers() ([]User, error) {
	rows, err := s.db.Query(`
SELECT id, name, client_private_key, client_public_key, assigned_ip, active, created_at, revoked_at
FROM users ORDER BY id ASC
`)
	if err != nil {
		return nil, fmt.Errorf("select users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// ListActiveUsers — только активные peers.
func (s *Store) ListActiveUsers() ([]User, error) {
	rows, err := s.db.Query(`
SELECT id, name, client_private_key, client_public_key, assigned_ip, active, created_at, revoked_at
FROM users WHERE active = 1 ORDER BY id ASC
`)
	if err != nil {
		return nil, fmt.Errorf("select active users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// GetUserByName ищет пользователя по имени.
func (s *Store) GetUserByName(name string) (User, error) {
	row := s.db.QueryRow(`
SELECT id, name, client_private_key, client_public_key, assigned_ip, active, created_at, revoked_at
FROM users WHERE name = ?
`, name)
	return scanUser(row)
}

// RevokeUser помечает пользователя неактивным.
func (s *Store) RevokeUser(name string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.Exec(`
UPDATE users SET active = 0, revoked_at = ? WHERE name = ? AND active = 1
`, now, name)
	if err != nil {
		return fmt.Errorf("revoke user: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("user %q not found or already revoked", name)
	}
	return nil
}

// NextClientIP выделяет следующий свободный IP в подсети (10.99.0.2, .3, ...).
func (s *Store) NextClientIP(subnetCIDR string) (string, error) {
	_, ipNet, err := net.ParseCIDR(subnetCIDR)
	if err != nil {
		return "", fmt.Errorf("parse subnet: %w", err)
	}

	base := ipNet.IP.To4()
	if base == nil {
		return "", fmt.Errorf("only IPv4 subnets supported")
	}

	rows, err := s.db.Query(`SELECT assigned_ip FROM users`)
	if err != nil {
		return "", fmt.Errorf("select assigned_ip: %w", err)
	}
	defer rows.Close()

	used := map[string]struct{}{}
	for rows.Next() {
		var ip string
		if err := rows.Scan(&ip); err != nil {
			return "", err
		}
		used[ip] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return "", err
	}

	// .1 — адрес сервера в нашей схеме; клиенты с .2
	for host := 2; host < 254; host++ {
		candidate := fmt.Sprintf("%d.%d.%d.%d", base[0], base[1], base[2], host)
		if _, ok := used[candidate]; !ok {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("no free IPs in subnet %s", subnetCIDR)
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanUser(scanner rowScanner) (User, error) {
	var (
		u         User
		active    int
		createdAt string
		revokedAt sql.NullString
	)
	err := scanner.Scan(
		&u.ID,
		&u.Name,
		&u.ClientPrivateKey,
		&u.ClientPublicKey,
		&u.AssignedIP,
		&active,
		&createdAt,
		&revokedAt,
	)
	if err == sql.ErrNoRows {
		return User{}, fmt.Errorf("user not found")
	}
	if err != nil {
		return User{}, fmt.Errorf("scan user: %w", err)
	}

	u.Active = active == 1
	t, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return User{}, err
	}
	u.CreatedAt = t

	if revokedAt.Valid {
		rt, err := time.Parse(time.RFC3339, revokedAt.String)
		if err != nil {
			return User{}, err
		}
		u.RevokedAt = &rt
	}
	return u, nil
}
