package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var userListCmd = &cobra.Command{
	Use:   "list",
	Short: "Список пользователей",
	RunE:  runUserList,
}

func init() {
	userCmd.AddCommand(userListCmd)
}

func runUserList(_ *cobra.Command, _ []string) error {
	s := mustStore()
	defer s.Close()

	users, err := s.ListUsers()
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tIP\tACTIVE\tCREATED")
	for _, u := range users {
		active := "yes"
		if !u.Active {
			active = "no"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", u.Name, u.AssignedIP, active, u.CreatedAt.Format("2006-01-02"))
	}
	return w.Flush()
}
