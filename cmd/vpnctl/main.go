// vpnctl — CLI для управления WireGuard-пользователями.
//
// Точка входа в Go-приложение всегда в package main, функция main().
// Аналог if __name__ == "__main__" в Python.
package main

import (
	"fmt"
	"os"

	"github.com/Muaytai/vpn-manager/cmd/vpnctl/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
