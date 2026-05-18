package configgen

import (
	"strings"
	"testing"
	"time"

	"github.com/Muaytai/vpn-manager/internal/store"
)

func TestClientConfigAdminOnly(t *testing.T) {
	s := store.ServerSettings{
		ListenPort:       51821,
		SubnetCIDR:       "10.99.0.0/24",
		ServerPublicKey:  "serverPubKeyBase64=",
		EndpointHost:     "203.0.113.10",
	}
	u := store.User{
		Name:             "alice",
		ClientPrivateKey: "clientPriv=",
		AssignedIP:       "10.99.0.2",
	}

	out := ClientConfig(s, u, ClientOptions{AdminOnly: true})
	if !strings.Contains(out, "AllowedIPs = 10.99.0.1/32") {
		t.Fatalf("admin-only must restrict routes, got:\n%s", out)
	}
	if strings.Contains(out, "0.0.0.0/0") {
		t.Fatal("admin-only must not include default route")
	}
}

func TestFullServerConfigIncludesPeer(t *testing.T) {
	s := store.ServerSettings{
		ListenPort:       51821,
		SubnetCIDR:       "10.99.0.0/24",
		ServerPrivateKey: "srvPriv=",
	}
	users := []store.User{{
		Name:             "bob",
		ClientPublicKey:  "bobPub=",
		AssignedIP:       "10.99.0.3",
		Active:           true,
		CreatedAt:        time.Now(),
	}}

	out := FullServerConfig(s, users)
	if !strings.Contains(out, "ListenPort = 51821") {
		t.Fatal("missing interface section")
	}
	if !strings.Contains(out, "bobPub=") {
		t.Fatal("missing peer section")
	}
}
