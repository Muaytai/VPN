package cmd

import (
	"github.com/spf13/cobra"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Управление VPN-пользователями",
}

func init() {
	rootCmd.AddCommand(userCmd)
}
