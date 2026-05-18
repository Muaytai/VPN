package paths

import (
	"os"
	"path/filepath"
)

// DataDir — каталог для БД и сгенерированных конфигов.
// Переопределяется переменной окружения VPN_DATA_DIR.
func DataDir() string {
	if v := os.Getenv("VPN_DATA_DIR"); v != "" {
		return v
	}
	return "./data"
}

func DBPath() string {
	return filepath.Join(DataDir(), "vpn.db")
}

func ClientConfigsDir() string {
	return filepath.Join(DataDir(), "client-configs")
}
