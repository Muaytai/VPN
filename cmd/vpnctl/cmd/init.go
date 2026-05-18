package cmd

import (
	"fmt"
	"time"

	"github.com/Muaytai/vpn-manager/internal/paths"
	"github.com/Muaytai/vpn-manager/internal/store"
	"github.com/Muaytai/vpn-manager/internal/wgkeys"
	"github.com/spf13/cobra"
)

var (
	initEndpoint string
	initPort     int
	initSubnet   string
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Инициализировать БД и ключи сервера (один раз)",
	Long: `Создаёт SQLite и пару ключей сервера WireGuard.

Пример:
  vpnctl init --endpoint 188.137.180.149 --port 51821`,
	RunE: runInit,
}

func init() {
	initCmd.Flags().StringVar(&initEndpoint, "endpoint", "", "Публичный IP или домен сервера (для клиентских .conf)")
	initCmd.Flags().IntVar(&initPort, "port", 51821, "UDP-порт WireGuard (51821 — безопасный выбор рядом с Docker)")
	initCmd.Flags().StringVar(&initSubnet, "subnet", "10.99.0.0/24", "Подсеть VPN (не пересекается с docker0 по умолчанию)")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, _ []string) error {
	s, err := openStore()
	if err != nil {
		return err
	}
	defer s.Close()

	ok, err := s.HasServerSettings()
	if err != nil {
		return err
	}
	if ok {
		return fmt.Errorf("already initialized — delete %s only if you know what you are doing", paths.DBPath())
	}

	serverKeys, err := wgkeys.Generate()
	if err != nil {
		return err
	}

	settings := store.ServerSettings{
		ListenPort:       initPort,
		SubnetCIDR:       initSubnet,
		ServerPrivateKey: serverKeys.Private,
		ServerPublicKey:  serverKeys.Public,
		EndpointHost:     initEndpoint,
		CreatedAt:        time.Now().UTC(),
	}

	if err := s.SaveServerSettings(settings); err != nil {
		return err
	}

	fmt.Println("OK: VPN manager initialized")
	fmt.Printf("DB: %s\n", paths.DBPath())
	fmt.Printf("Server public key: %s\n", serverKeys.Public)
	fmt.Printf("Listen port: %d\n", initPort)
	fmt.Printf("Subnet: %s\n", initSubnet)
	if initEndpoint == "" {
		fmt.Println("\nПодсказка: задай --endpoint при init или обнови вручную в БД перед выдачей клиентам.")
	}
	fmt.Println("\nСледующий шаг: vpnctl user add alice")
	return nil
}
