package cmd

import (
	"fmt"
	"os"

	"github.com/Muaytai/vpn-manager/internal/configgen"
	"github.com/Muaytai/vpn-manager/internal/netutil"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Конфигурация WireGuard на сервере",
}

var (
	renderWAN string
)

var serverRenderCmd = &cobra.Command{
	Use:   "render",
	Short: "Вывести wg0.conf (full-tunnel + NAT, если включено в БД)",
	RunE:  runServerRender,
}

var serverApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Включить ip_forward, записать wg0.conf и применить WireGuard",
	Long: `Только на Linux-сервере. Перед apply сделай бэкап iptables.

  export VPN_DATA_DIR=/opt/vpn-manager/data
  vpnctl server apply`,
	RunE: runServerApply,
}

var serverConfigureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Обновить режим full-tunnel / WAN / DNS в БД",
	RunE:  runServerConfigure,
}

var (
	cfgWAN        string
	cfgFullTunnel bool
	cfgDNS        string
)

func init() {
	serverRenderCmd.Flags().StringVar(&renderWAN, "wan", "", "WAN-интерфейс (перекрывает значение из БД)")
	serverConfigureCmd.Flags().StringVar(&cfgWAN, "wan", "", "WAN-интерфейс (ens3, eth0)")
	serverConfigureCmd.Flags().BoolVar(&cfgFullTunnel, "full-tunnel", true, "Весь интернет через VPN")
	serverConfigureCmd.Flags().StringVar(&cfgDNS, "dns", "1.1.1.1, 8.8.8.8", "DNS для клиентов")

	serverCmd.AddCommand(serverRenderCmd, serverApplyCmd, serverConfigureCmd)
	rootCmd.AddCommand(serverCmd)
}

func runServerRender(_ *cobra.Command, _ []string) error {
	conf, err := buildServerConfig(renderWAN)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(os.Stdout, conf)
	return err
}

func runServerApply(_ *cobra.Command, _ []string) error {
	conf, wan, err := buildServerConfigWithWAN(renderWAN)
	if err != nil {
		return err
	}

	if err := netutil.EnableIPForward(); err != nil {
		return fmt.Errorf("enable ip_forward: %w", err)
	}

	const wgConfig = "/etc/wireguard/wg0.conf"
	if err := os.WriteFile(wgConfig, []byte(conf), 0o600); err != nil {
		return fmt.Errorf("write %s: %w", wgConfig, err)
	}

	fmt.Printf("OK: wrote %s (WAN=%s)\n", wgConfig, wan)

	if err := netutil.SyncWireGuard(); err != nil {
		fmt.Println("wg syncconf failed, try: systemctl restart wg-quick@wg0")
		fmt.Println(err)
		return nil
	}
	fmt.Println("OK: wg syncconf applied")
	fmt.Println("\nПроверка: wg show")
	fmt.Println("Клиентам: vpnctl user regen <name> и переимпорт .conf в WireGuard")
	return nil
}

func runServerConfigure(_ *cobra.Command, _ []string) error {
	s, err := openStore()
	if err != nil {
		return err
	}
	defer s.Close()

	if err := s.UpdateServerNetwork(cfgFullTunnel, cfgWAN, cfgDNS); err != nil {
		return err
	}
	fmt.Printf("OK: full_tunnel=%v wan=%q dns=%q\n", cfgFullTunnel, cfgWAN, cfgDNS)
	fmt.Println("Дальше: vpnctl user regen alice  &&  vpnctl server apply")
	return nil
}

func buildServerConfig(wanOverride string) (string, error) {
	conf, _, err := buildServerConfigWithWAN(wanOverride)
	return conf, err
}

func buildServerConfigWithWAN(wanOverride string) (string, string, error) {
	s, err := openStore()
	if err != nil {
		return "", "", err
	}
	defer s.Close()

	settings, err := s.GetServerSettings()
	if err != nil {
		return "", "", err
	}

	users, err := s.ListActiveUsers()
	if err != nil {
		return "", "", err
	}

	wan := wanOverride
	if wan == "" {
		wan = settings.WANInterface
	}
	if settings.FullTunnel {
		resolved, err := netutil.ResolveWANInterface(wan)
		if err != nil {
			return "", "", fmt.Errorf("WAN interface: %w (укажи: vpnctl server configure --wan ens3)", err)
		}
		wan = resolved
	}

	opts := configgen.ServerRenderOptions{WANInterface: wan}
	if !settings.FullTunnel {
		opts.WANInterface = ""
	}

	conf := configgen.FullServerConfig(settings, users, opts)
	return conf, wan, nil
}
