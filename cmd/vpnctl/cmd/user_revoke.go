package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var userRevokeCmd = &cobra.Command{
	Use:   "revoke [name]",
	Short: "Отозвать доступ пользователя",
	Args:  cobra.ExactArgs(1),
	RunE:  runUserRevoke,
}

func init() {
	userCmd.AddCommand(userRevokeCmd)
}

func runUserRevoke(_ *cobra.Command, args []string) error {
	name := args[0]

	s, err := openStore()
	if err != nil {
		return err
	}
	defer s.Close()

	if err := s.RevokeUser(name); err != nil {
		return err
	}

	fmt.Printf("OK: user %q revoked\n", name)
	fmt.Println("Дальше: vpnctl server apply")
	return nil
}
