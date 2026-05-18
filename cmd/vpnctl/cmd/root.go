package cmd

import (
	"fmt"
	"os"

	"github.com/Muaytai/vpn-manager/internal/paths"
	"github.com/Muaytai/vpn-manager/internal/store"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "vpnctl",
	Short: "Управление WireGuard VPN (учебный проект на Go)",
	Long: `vpnctl — обёртка вокруг WireGuard.

Мы НЕ пишем свой VPN-протокол. WireGuard работает в ядре Linux,
а vpnctl хранит пользователей, ключи и генерирует .conf файлы.`,
}

// Execute запускает CLI (вызывается из main).
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(func() {
		// Можно добавить чтение .env позже
	})
}

func openStore() (*store.Store, error) {
	return store.Open(paths.DBPath())
}

func mustStore() *store.Store {
	s, err := openStore()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return s
}
