package cmd

import (
	"fmt"
	"os"

	"github.com/Muaytai/vpn-manager/internal/configgen"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Конфигурация WireGuard на сервере",
}

var serverRenderCmd = &cobra.Command{
	Use:   "render",
	Short: "Вывести полный wg0.conf для активных пользователей",
	Long: `Печатает готовый wg0.conf в stdout.

На сервере:
  vpnctl server render | sudo tee /etc/wireguard/wg0.conf`,
	RunE: runServerRender,
}

func init() {
	serverCmd.AddCommand(serverRenderCmd)
	rootCmd.AddCommand(serverCmd)
}

func runServerRender(_ *cobra.Command, _ []string) error {
	s, err := openStore()
	if err != nil {
		return err
	}
	defer s.Close()

	settings, err := s.GetServerSettings()
	if err != nil {
		return err
	}

	users, err := s.ListActiveUsers()
	if err != nil {
		return err
	}

	conf := configgen.FullServerConfig(settings, users)
	_, err = fmt.Fprint(os.Stdout, conf)
	return err
}
