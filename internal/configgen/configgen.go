// Package configgen формирует текстовые .conf для WireGuard.
package configgen

import (
	"fmt"
	"strings"

	"github.com/Muaytai/vpn-manager/internal/store"
)

// ClientOptions — параметры клиентского конфига.
type ClientOptions struct {
	// AdminOnly: true = только VPN-IP сервера (без выхода в интернет через VPN).
	AdminOnly bool
	DNS       string
}

// ServerRenderOptions — параметры wg0.conf на сервере.
type ServerRenderOptions struct {
	WANInterface string
}

// ServerInterfaceConfig — секция [Interface] для wg0 на сервере.
func ServerInterfaceConfig(s store.ServerSettings, opts ServerRenderOptions) string {
	var b strings.Builder
	fmt.Fprintf(&b, "[Interface]\n")
	fmt.Fprintf(&b, "Address = %s\n", serverAddressCIDR(s.SubnetCIDR))
	fmt.Fprintf(&b, "ListenPort = %d\n", s.ListenPort)
	fmt.Fprintf(&b, "PrivateKey = %s\n", s.ServerPrivateKey)

	if s.FullTunnel && opts.WANInterface != "" {
		iface := opts.WANInterface
		// -I в начало цепочек: на хосте с Docker FORWARD policy DROP, wg syncconf не вызывает PostUp.
		fmt.Fprintf(&b, "PostUp = iptables -I DOCKER-USER 1 -i wg0 -j ACCEPT; iptables -I DOCKER-USER 1 -o wg0 -j ACCEPT; iptables -I FORWARD 1 -i wg0 -j ACCEPT; iptables -I FORWARD 1 -o wg0 -j ACCEPT; iptables -t nat -A POSTROUTING -o %s -j MASQUERADE\n", iface)
		fmt.Fprintf(&b, "PostDown = iptables -D DOCKER-USER -i wg0 -j ACCEPT; iptables -D DOCKER-USER -o wg0 -j ACCEPT; iptables -D FORWARD -i wg0 -j ACCEPT; iptables -D FORWARD -o wg0 -j ACCEPT; iptables -t nat -D POSTROUTING -o %s -j MASQUERADE\n", iface)
	}
	return b.String()
}

// PeerBlock — секция [Peer] для одного пользователя.
func PeerBlock(u store.User) string {
	return fmt.Sprintf(
		"[Peer]\nPublicKey = %s\nAllowedIPs = %s/32\n",
		u.ClientPublicKey,
		u.AssignedIP,
	)
}

// ClientConfig — полный конфиг для телефона/ПК.
func ClientConfig(s store.ServerSettings, u store.User, opts ClientOptions) string {
	// Только IPv4: ::/0 на Windows часто ломает DNS и «убивает» интернет.
	allowedIPs := "0.0.0.0/0"
	if opts.AdminOnly {
		allowedIPs = serverAddressCIDR(s.SubnetCIDR)
		if !strings.Contains(allowedIPs, "/") {
			allowedIPs += "/32"
		}
	}

	dns := normalizeDNS(opts.DNS, s.DefaultDNS)

	var b strings.Builder
	fmt.Fprintf(&b, "[Interface]\n")
	fmt.Fprintf(&b, "PrivateKey = %s\n", u.ClientPrivateKey)
	fmt.Fprintf(&b, "Address = %s/32\n", u.AssignedIP)
	if !opts.AdminOnly {
		fmt.Fprintf(&b, "MTU = 1420\n")
	}
	if dns != "" && !opts.AdminOnly {
		fmt.Fprintf(&b, "DNS = %s\n", dns)
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
func FullServerConfig(s store.ServerSettings, users []store.User, opts ServerRenderOptions) string {
	var b strings.Builder
	b.WriteString(ServerInterfaceConfig(s, opts))
	for _, u := range users {
		if !u.Active {
			continue
		}
		b.WriteString("\n")
		b.WriteString(PeerBlock(u))
	}
	return b.String()
}

func serverAddressCIDR(subnetCIDR string) string {
	// 10.99.0.0/24 -> 10.99.0.1/24
	parts := strings.Split(subnetCIDR, "/")
	prefix := "24"
	if len(parts) == 2 && parts[1] != "" {
		prefix = parts[1]
	}
	ip := parts[0]
	octets := strings.Split(ip, ".")
	if len(octets) == 4 {
		octets[3] = "1"
		return strings.Join(octets, ".") + "/" + prefix
	}
	return ip + "/" + prefix
}

// normalizeDNS — один адрес DNS (WireGuard на Windows плохо переваривает "1.1.1.1, 8.8.8.8").
func normalizeDNS(override, fallback string) string {
	dns := strings.TrimSpace(override)
	if dns == "" {
		dns = strings.TrimSpace(fallback)
	}
	if dns == "" {
		return "1.1.1.1"
	}
	if idx := strings.IndexAny(dns, ",;"); idx >= 0 {
		dns = strings.TrimSpace(dns[:idx])
	}
	return dns
}
