#!/usr/bin/env bash
# Установка на Ubuntu-сервер (НЕ трогает Docker-контейнеры).
set -euo pipefail

INSTALL_DIR="${INSTALL_DIR:-/opt/vpn-manager}"
DATA_DIR="${DATA_DIR:-/var/lib/vpn-manager}"

if [[ "$(id -u)" -ne 0 ]]; then
  echo "Запусти от root: sudo bash deploy/install.sh"
  exit 1
fi

echo "==> Установка wireguard-tools (если ещё нет)"
apt-get update -qq
apt-get install -y wireguard-tools

echo "==> Каталоги"
mkdir -p "$INSTALL_DIR/bin" "$DATA_DIR/client-configs" /etc/wireguard
chmod 700 "$DATA_DIR"

echo "==> Копирование бинарника (собери локально: go build -o bin/vpnctl ./cmd/vpnctl)"
if [[ ! -f ./bin/vpnctl ]]; then
  echo "Ошибка: нет ./bin/vpnctl — сначала собери проект"
  exit 1
fi
install -m 755 ./bin/vpnctl "$INSTALL_DIR/bin/vpnctl"

echo "==> systemd unit"
install -m 644 ./deploy/vpn-manager.service /etc/systemd/system/vpn-manager.service
systemctl daemon-reload

echo ""
echo "Готово. Дальше на сервере:"
echo "  export VPN_DATA_DIR=$DATA_DIR"
echo "  $INSTALL_DIR/bin/vpnctl init --endpoint YOUR_IP --port 51821 --wan ens3"
echo "  $INSTALL_DIR/bin/vpnctl user add alice"
echo "  iptables-save > /root/iptables.backup.txt"
echo "  $INSTALL_DIR/bin/vpnctl server apply"
echo "  systemctl enable --now wg-quick@wg0"
echo ""
echo "Full-tunnel: весь интернет через VPN. Проверь Docker после apply."
