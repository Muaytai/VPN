// Package configgen формирует текстовые .conf для WireGuard.
package configgen

import (
	"fmt"
	"strings"

	"github.com/Muaytai/vpn-manager/internal/store"
)

// ClientOptions — параметры клиентского конфига.
type ClientOptions struct {
	// AdminOnly: true = доступ только к VPN-IP сервера (безопасный режим для Docker-хоста).
	AdminOnly bool
	DNS       string
}

// ServerInterfaceConfig — секция [Interface] для wg0 на сервере.
func ServerInterfaceConfig(s store.ServerSettings) string {
	var b strings.Builder
	fmt.Fprintf(&b, "[Interface]\n")
	fmt.Fprintf(&b, "Address = %s\n", serverAddress(s.SubnetCIDR))
	fmt.Fprintf(&b, "ListenPort = %d\n", s.ListenPort)
	fmt.Fprintf(&b, "PrivateKey = %s\n", s.ServerPrivateKey)
	fmt.Fprintf(&b, "\n# PostUp/PostDown с iptables добавляй только когда нужен full-tunnel.\n")
	fmt.Fprintf(&b, "# На боевом сервере с Docker сначала используй AdminOnly у клиентов.\n")
	return b.String()
}

// PeerBlock — секция [Peer] для одного пользователя (вставляется в wg0.conf).
func PeerBlock(u store.User) string {
	return fmt.Sprintf(
		"[Peer]\nPublicKey = %s\nAllowedIPs = %s/32\n",
		u.ClientPublicKey,
		u.AssignedIP,
	)
}

// ClientConfig — полный конфиг для телефона/ПК.
func ClientConfig(s store.ServerSettings, u store.User, opts ClientOptions) string {
	allowedIPs := "0.0.0.0/0, ::/0"
	if opts.AdminOnly {
		allowedIPs = serverAddress(s.SubnetCIDR) + "/32"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "[Interface]\n")
	fmt.Fprintf(&b, "PrivateKey = %s\n", u.ClientPrivateKey)
	fmt.Fprintf(&b, "Address = %s/32\n", u.AssignedIP)
	if opts.DNS != "" && !opts.AdminOnly {
		fmt.Fprintf(&b, "DNS = %s\n", opts.DNS)
	}
	fmt.Fprintf(&b, "\n[Peer]\n")
	fmt.Fprintf(&b, "PublicKey = %s\n", s.ServerPublicKey)
	if s.EndpointHost != "" {
		fmt.Fprintf(&b, "Endpoint = %s:%d\n", s.EndpointHost, s.ListenPort)
	}
	fmt.Fprintf(&b, "AllowedIPs = %s\n", allowedIPs)
	fmt.Fprintf(&b, "PersistentKeepalive = 25\n")
	return b.String()
}

// FullServerConfig собирает wg0.conf со всеми активными peers.
func FullServerConfig(s store.ServerSettings, users []store.User) string {
	var b strings.Builder
	b.WriteString(ServerInterfaceConfig(s))
	for _, u := range users {
		if !u.Active {
			continue
		}
		b.WriteString("\n")
		b.WriteString(PeerBlock(u))
	}
	return b.String()
}

func serverAddress(subnetCIDR string) string {
	// 10.99.0.0/24 -> 10.99.0.1
	parts := strings.Split(subnetCIDR, "/")
	ip := parts[0]
	octets := strings.Split(ip, ".")
	if len(octets) == 4 {
		octets[3] = "1"
		return strings.Join(octets, ".")
	}
	return ip
}
