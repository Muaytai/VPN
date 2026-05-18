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
	initEndpoint    string
	initPort        int
	initSubnet      string
	initWAN         string
	initFullTunnel  bool
	initDNS         string
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Инициализировать БД и ключи сервера (один раз)",
	Long: `Создаёт SQLite и пару ключей сервера WireGuard.

По умолчанию: full-tunnel (весь интернет через VPN) + NAT на сервере.

Пример:
  vpnctl init --endpoint 188.137.180.149 --port 51821 --wan ens3`,
	RunE: runInit,
}

func init() {
	initCmd.Flags().StringVar(&initEndpoint, "endpoint", "", "Публичный IP или домен сервера")
	initCmd.Flags().IntVar(&initPort, "port", 51821, "UDP-порт WireGuard")
	initCmd.Flags().StringVar(&initSubnet, "subnet", "10.99.0.0/24", "Подсеть VPN")
	initCmd.Flags().StringVar(&initWAN, "wan", "", "WAN-интерфейс (ens3, eth0). Пусто = авто при server apply")
	initCmd.Flags().BoolVar(&initFullTunnel, "full-tunnel", true, "Весь интернет через VPN (NAT на сервере)")
	initCmd.Flags().StringVar(&initDNS, "dns", "1.1.1.1, 8.8.8.8", "DNS в клиентских .conf")
	rootCmd.AddCommand(initCmd)
}

func runInit(_ *cobra.Command, _ []string) error {
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
		FullTunnel:       initFullTunnel,
		WANInterface:     initWAN,
		DefaultDNS:       initDNS,
		CreatedAt:        time.Now().UTC(),
	}

	if err := s.SaveServerSettings(settings); err != nil {
		return err
	}

	fmt.Println("OK: VPN manager initialized (full-tunnel mode)")
	fmt.Printf("DB: %s\n", paths.DBPath())
	fmt.Printf("Server public key: %s\n", serverKeys.Public)
	fmt.Printf("Full tunnel: %v\n", initFullTunnel)
	fmt.Printf("DNS for clients: %s\n", initDNS)
	if initEndpoint == "" {
		fmt.Println("\nПодсказка: задай --endpoint (IP сервера) перед выдачей конфигов.")
	}
	fmt.Println("\nСледующий шаг: vpnctl user add alice")
	fmt.Println("На сервере: vpnctl server apply")
	return nil
}
