// Package netutil — сетевые утилиты для Linux-сервера.
package netutil

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// DefaultWANInterface возвращает интерфейс маршрута по умолчанию (eth0, ens3, …).
func DefaultWANInterface() (string, error) {
	out, err := exec.Command("ip", "-4", "route", "show", "default").Output()
	if err != nil {
		return "", fmt.Errorf("ip route: %w", err)
	}
	line := strings.TrimSpace(string(out))
	if line == "" {
		return "", fmt.Errorf("no default route")
	}
	fields := strings.Fields(line)
	// default via 1.2.3.4 dev ens3 ...
	for i := 0; i < len(fields)-1; i++ {
		if fields[i] == "dev" {
			return fields[i+1], nil
		}
	}
	return "", fmt.Errorf("parse default route: %q", line)
}

// ResolveWANInterface — явное имя или автоопределение.
func ResolveWANInterface(configured string) (string, error) {
	if configured != "" {
		return configured, nil
	}
	return DefaultWANInterface()
}

// IPForwardEnabled проверяет net.ipv4.ip_forward=1.
func IPForwardEnabled() (bool, error) {
	out, err := exec.Command("sysctl", "-n", "net.ipv4.ip_forward").Output()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) == "1", nil
}

// EnableIPForward включает пересылку IPv4 (нужно для NAT).
func EnableIPForward() error {
	if err := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1").Run(); err != nil {
		return err
	}
	return exec.Command("sh", "-c", `grep -qs '^net.ipv4.ip_forward=1' /etc/sysctl.d/99-vpn-manager.conf || echo 'net.ipv4.ip_forward=1' >> /etc/sysctl.d/99-vpn-manager.conf`).Run()
}

// SyncWireGuard применяет wg0.conf без полного рестарта (Linux).
func SyncWireGuard() error {
	cmd := exec.Command("bash", "-c", `wg syncconf wg0 <(wg-quick strip wg0)`)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("wg syncconf: %w: %s", err, stderr.String())
	}
	return nil
}
