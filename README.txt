================================================================================
  VPN Manager (vpnctl) — инструкция администратора
  Репозиторий: https://github.com/Muaytai/VPN
  Сервер: 188.137.180.149  |  Порт WireGuard: UDP 51821
================================================================================

ЧТО ЭТО
-------
WireGuard на VPS + программа vpnctl (Go) для пользователей и .conf файлов.
Мы НЕ пишем свой VPN-протокол — шифрование делает WireGuard в ядре Linux.

Режим по умолчанию: FULL-TUNNEL (весь интернет через VPN, как у alice).
Внешний IP у всех: 188.137.180.149 (Нидерланды, Zomro).


ПУТИ НА СЕРВЕРЕ
---------------
  Код:     /opt/vpn-manager
  Данные:  /opt/vpn-manager/data
  БД:      /opt/vpn-manager/data/vpn.db
  Конфиги: /opt/vpn-manager/data/client-configs/
  wg0:     /etc/wireguard/wg0.conf

Перед каждой сессией:
  export VPN_DATA_DIR=/opt/vpn-manager/data
  cd /opt/vpn-manager


================================================================================
  ДОБАВИТЬ ПОЛЬЗОВАТЕЛЯ
================================================================================

  /opt/vpn-manager/bin/vpnctl user add ИМЯ
  /opt/vpn-manager/bin/vpnctl server apply
  /opt/vpn-manager/bin/vpnctl user list

  Конфиг для отправки в Telegram:
    /opt/vpn-manager/data/client-configs/ИМЯ.conf

  Имя — латиница, без @ (пример: @ivan_petrov → ivan_petrov).

  Только VPN-сеть сервера, БЕЗ полного интернета (редко):
    /opt/vpn-manager/bin/vpnctl user add ИМЯ --admin-only
    /opt/vpn-manager/bin/vpnctl server apply


================================================================================
  УДАЛИТЬ / ОТОЗВАТЬ ДОСТУП
================================================================================

  /opt/vpn-manager/bin/vpnctl user revoke ИМЯ
  /opt/vpn-manager/bin/vpnctl server apply
  /opt/vpn-manager/bin/vpnctl user list

  Старый .conf после revoke перестаёт работать.
  Запись в БД остаётся (active = no).


================================================================================
  ОБНОВИТЬ КОНФИГ ПОЛЬЗОВАТЕЛЯ
================================================================================

  /opt/vpn-manager/bin/vpnctl user regen ИМЯ

  server apply после regen НЕ нужен.
  Человеку — переимпорт .conf в WireGuard.


================================================================================
  КОГДА НУЖЕН server apply
================================================================================

  ДА:   user add, user revoke, server configure
  НЕТ:  user regen, user list


================================================================================
  ДРУГИЕ КОМАНДЫ
================================================================================

  Список пользователей:
    /opt/vpn-manager/bin/vpnctl user list

  Статус WireGuard:
    wg show

  Настройки full-tunnel / WAN / DNS:
    ip route | awk '/default/ {print $5}'
    /opt/vpn-manager/bin/vpnctl server configure --wan ens3 --full-tunnel --dns 1.1.1.1
    /opt/vpn-manager/bin/vpnctl server apply

  Обновить код с GitHub:
    cd /opt/vpn-manager
    git pull
    /usr/local/go/bin/go build -o bin/vpnctl ./cmd/vpnctl
    export VPN_DATA_DIR=/opt/vpn-manager/data
    /opt/vpn-manager/bin/vpnctl server apply


================================================================================
  КОМАНДА (Telegram → файл конфига)
================================================================================

  @egrchik1        →  egrchik1.conf        (10.99.0.3)
  @tatsiana_tg     →  tatsiana_tg.conf     (10.99.0.4)
  @DzmitryKosmach  →  dzmitrykosmach.conf  (10.99.0.5)
  @meloman1565     →  meloman1565.conf     (10.99.0.6)
  @doc_chaizer     →  doc_chaizer.conf     (10.99.0.7)
  @Divolker        →  divolker.conf        (10.99.0.8)
  @TataKach        →  tatakach.conf        (10.99.0.9)
  alice (админ)    →  alice.conf           (10.99.0.2)

  У всех: полный интернет через VPN (как у alice).
  Отправлять .conf только в ЛИЧКУ, не в общий чат!


================================================================================
  ИНСТРУКЦИЯ ДЛЯ СОТРУДНИКОВ (можно копировать в Telegram)
================================================================================

  1. Установи WireGuard:
     https://www.wireguard.com/install/

  2. Импортируй присланный файл .conf

  3. Включи туннель

  4. Проверка: открой https://ifconfig.me
     Должен быть IP: 188.137.180.149


================================================================================
  ЛОКАЛЬНО НА WINDOWS (разработка)
================================================================================

  cd D:\Работа\VPN
  $env:VPN_DATA_DIR = ".\data"

  go run .\cmd\vpnctl user add ИМЯ
  go run .\cmd\vpnctl user revoke ИМЯ
  go run .\cmd\vpnctl user list

  После изменений на ПК — скопируй папку data на сервер в
  /opt/vpn-manager/data/ и выполни server apply.
  Или делай add/revoke сразу на сервере (удобнее).


================================================================================
  БЭКАП И ОТКАТ (если сломался интернет / Docker)
================================================================================

  Перед server apply:
    iptables-save > /root/iptables.backup.txt

  Остановить VPN:
    systemctl stop wg-quick@wg0

  Откат iptables:
    iptables-restore < /root/iptables.backup.txt


================================================================================
  ВАЖНО
================================================================================

  - Папку data/ и .conf НЕ коммитить в Git (приватные ключи).
  - После каждого add/revoke — server apply на сервере.
  - Сервер с Docker: NAT и DOCKER-USER прописаны в wg0.conf (PostUp).
  - Gemini может не открыться у части VPS-IP — это ограничение Google, не баг VPN.

================================================================================
