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

// RestartWireGuard перезапускает wg0 и выполняет PostUp/PostDown из wg0.conf.
// wg syncconf НЕ запускает PostUp — на Docker-хосте без restart нет NAT.
func RestartWireGuard() error {
	if err := exec.Command("systemctl", "restart", "wg-quick@wg0").Run(); err != nil {
		if err2 := exec.Command("wg-quick", "down", "wg0").Run(); err2 != nil {
			_ = err2
		}
		if err3 := exec.Command("wg-quick", "up", "wg0").Run(); err3 != nil {
			return fmt.Errorf("wg-quick up: %w", err3)
		}
	}
	return nil
}
