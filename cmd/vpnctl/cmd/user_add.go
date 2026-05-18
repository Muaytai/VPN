package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Muaytai/vpn-manager/internal/configgen"
	"github.com/Muaytai/vpn-manager/internal/paths"
	"github.com/Muaytai/vpn-manager/internal/store"
	"github.com/Muaytai/vpn-manager/internal/wgkeys"
	"github.com/spf13/cobra"
)

var (
	addAdminOnly bool
	addDNS       string
)

var userAddCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Добавить пользователя и сгенерировать клиентский .conf",
	Args:  cobra.ExactArgs(1),
	RunE:  runUserAdd,
}

func init() {
	userAddCmd.Flags().BoolVar(&addAdminOnly, "admin-only", true, "Безопасный режим: клиент видит только VPN-IP сервера")
	userAddCmd.Flags().StringVar(&addDNS, "dns", "1.1.1.1", "DNS для клиента (только если не --admin-only)")
	userCmd.AddCommand(userAddCmd)
}

func runUserAdd(_ *cobra.Command, args []string) error {
	name := args[0]

	s, err := openStore()
	if err != nil {
		return err
	}
	defer s.Close()

	settings, err := s.GetServerSettings()
	if err != nil {
		return err
	}

	clientKeys, err := wgkeys.Generate()
	if err != nil {
		return err
	}

	ip, err := s.NextClientIP(settings.SubnetCIDR)
	if err != nil {
		return err
	}

	user := store.User{
		Name:             name,
		ClientPrivateKey: clientKeys.Private,
		ClientPublicKey:  clientKeys.Public,
		AssignedIP:       ip,
		CreatedAt:        time.Now().UTC(),
	}

	created, err := s.CreateUser(user)
	if err != nil {
		return err
	}

	confText := configgen.ClientConfig(settings, created, configgen.ClientOptions{
		AdminOnly: addAdminOnly,
		DNS:       addDNS,
	})

	if err := os.MkdirAll(paths.ClientConfigsDir(), 0o700); err != nil {
		return err
	}
	outPath := filepath.Join(paths.ClientConfigsDir(), name+".conf")
	if err := os.WriteFile(outPath, []byte(confText), 0o600); err != nil {
		return err
	}

	fmt.Println("OK: user created")
	fmt.Printf("Name: %s\n", created.Name)
	fmt.Printf("IP: %s\n", created.AssignedIP)
	fmt.Printf("Client config: %s\n", outPath)
	fmt.Println("\n--- Peer для wg0.conf на сервере ---")
	fmt.Print(configgen.PeerBlock(created))
	return nil
}
