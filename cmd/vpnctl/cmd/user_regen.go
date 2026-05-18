package cmd

import (
	"github.com/spf13/cobra"
)

var (
	regenAdminOnly bool
	regenDNS       string
)

var userRegenCmd = &cobra.Command{
	Use:   "regen [name]",
	Short: "Пересоздать клиентский .conf (после смены режима full-tunnel)",
	Args:  cobra.ExactArgs(1),
	RunE:  runUserRegen,
}

func init() {
	userRegenCmd.Flags().BoolVar(&regenAdminOnly, "admin-only", false, "Принудительно admin-only")
	userRegenCmd.Flags().StringVar(&regenDNS, "dns", "", "DNS (по умолчанию из настроек сервера)")
	userCmd.AddCommand(userRegenCmd)
}

func runUserRegen(_ *cobra.Command, args []string) error {
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

	user, err := s.GetUserByName(name)
	if err != nil {
		return err
	}
	if !user.Active {
		return errUserInactive(name)
	}

	adminOnly := regenAdminOnly
	if !regenAdminOnly {
		adminOnly = !settings.FullTunnel
	}

	if err := writeUserConfig(settings, user, adminOnly, regenDNS); err != nil {
		return err
	}
	return nil
}

func errUserInactive(name string) error {
	return &userInactiveError{name: name}
}

type userInactiveError struct{ name string }

func (e *userInactiveError) Error() string {
	return "user " + e.name + " is revoked"
}
