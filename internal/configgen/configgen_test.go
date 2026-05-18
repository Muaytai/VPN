package configgen

import (
	"strings"
	"testing"
	"time"

	"github.com/Muaytai/vpn-manager/internal/store"
)

func TestClientConfigFullTunnel(t *testing.T) {
	s := store.ServerSettings{
		ListenPort:      51821,
		SubnetCIDR:      "10.99.0.0/24",
		ServerPublicKey: "serverPub=",
		EndpointHost:    "203.0.113.10",
		FullTunnel:      true,
		DefaultDNS:      "1.1.1.1, 8.8.8.8",
	}
	u := store.User{
		Name:             "alice",
		ClientPrivateKey: "clientPriv=",
		AssignedIP:       "10.99.0.2",
	}

	out := ClientConfig(s, u, ClientOptions{AdminOnly: false})
	if !strings.Contains(out, "AllowedIPs = 0.0.0.0/0") {
		t.Fatalf("expected full tunnel routes, got:\n%s", out)
	}
	if !strings.Contains(out, "DNS = 1.1.1.1") {
		t.Fatal("expected DNS in full tunnel config")
	}
}

func TestClientConfigAdminOnly(t *testing.T) {
	s := store.ServerSettings{
		ListenPort:      51821,
		SubnetCIDR:      "10.99.0.0/24",
		ServerPublicKey: "serverPub=",
		EndpointHost:    "203.0.113.10",
	}
	u := store.User{
		ClientPrivateKey: "clientPriv=",
		AssignedIP:       "10.99.0.2",
	}

	out := ClientConfig(s, u, ClientOptions{AdminOnly: true})
	if !strings.Contains(out, "AllowedIPs = 10.99.0.1/24") {
		t.Fatalf("admin-only must restrict routes, got:\n%s", out)
	}
}

func TestFullServerConfigNAT(t *testing.T) {
	s := store.ServerSettings{
		ListenPort:       51821,
		SubnetCIDR:       "10.99.0.0/24",
		ServerPrivateKey: "srvPriv=",
		FullTunnel:       true,
	}
	users := []store.User{{
		ClientPublicKey: "bobPub=",
		AssignedIP:      "10.99.0.3",
		Active:          true,
		CreatedAt:       time.Now(),
	}}

	out := FullServerConfig(s, users, ServerRenderOptions{WANInterface: "ens3"})
	if !strings.Contains(out, "Address = 10.99.0.1/24") {
		t.Fatal("server address must include CIDR /24")
	}
	if !strings.Contains(out, "MASQUERADE") {
		t.Fatal("full tunnel server config must include NAT")
	}
}
