# VPN Manager 
Управление **WireGuard** через CLI: пользователи, ключи, генерация `.conf`.

> Мы **не** пишем свой VPN-протокол. WireGuard — в ядре Linux.  
> Наш код на Go — это «панель/скрипт», как админка на Python, только компилируется в один бинарник.

## Структура репозитория

```
vpn-manager/
├── cmd/vpnctl/           # точка входа CLI (как scripts/ или bin/)
│   ├── main.go
│   └── cmd/              # команды Cobra: init, user add, ...
├── internal/             # приватный код (Go не даст импортировать извне)
│   ├── store/            # SQLite, пользователи
│   ├── wgkeys/           # генерация ключей WireGuard
│   ├── configgen/        # шаблоны .conf
│   └── paths/            # пути к data/
├── deploy/               # установка на сервер (systemd)
├── configs/              # примеры wg0
├── data/                 # локально (в .gitignore)
└── go.mod
```

Пакет `internal/` — соглашение Go: «только для этого проекта», как приватные модули в monorepo.

---

## Быстрый старт (Windows)

### 1. Установи Go

https://go.dev/dl/ — версия **1.22+**.

Проверка:

```powershell
go version
```

### 2. Клонируй / открой папку проекта

```powershell
cd "d:\Работа\VPN"
go mod download
```

### 3. Сборка и тесты

```powershell
go test ./...
go build -o bin\vpnctl.exe .\cmd\vpnctl
```

### 4. Локальная инициализация (без сервера)

```powershell
$env:VPN_DATA_DIR = ".\data"
.\bin\vpnctl.exe init --endpoint 188.137.180.149 --port 51821
.\bin\vpnctl.exe user add alice
.\bin\vpnctl.exe user list
.\bin\vpnctl.exe server render
```

Файлы появятся в `data/`:
- `vpn.db` — SQLite
- `client-configs/alice.conf` — конфиг для WireGuard-клиента

Флаг `--admin-only` (по умолчанию **включён**) — клиент видит только IP VPN-сервера (`10.99.0.1/32`).  
Это **безопасный режим** для боевого сервера с Docker: не перенаправляет весь интернет через VPN.

---

## Команды CLI

| Команда | Что делает |
|---------|------------|
| `vpnctl init` | Создаёт БД и ключи сервера |
| `vpnctl user add NAME` | Новый пользователь + `.conf` |
| `vpnctl user list` | Список пользователей |
| `vpnctl user revoke NAME` | Отзыв доступа |
| `vpnctl server render` | Полный `wg0.conf` в stdout |

---

## Репозиторий GitHub

Аккаунт: [github.com/Muaytai](https://github.com/Muaytai)

После создания репозитория `vpn-manager` на GitHub:

```bash
git remote add origin https://github.com/Muaytai/vpn-manager.git
```

Модуль Go: `github.com/Muaytai/vpn-manager`

---

## Workflow: локально → GitHub → сервер

```
[Твой ПК]  git push   [GitHub]   git pull   [Сервер /opt/vpn-manager]
   │                                              │
   ├─ go test ./...                               ├─ go build
   ├─ vpnctl user add (dev)                       ├─ vpnctl init (один раз)
   └─ правки Go                                   ├─ vpnctl server render → wg0.conf
                                                  └─ systemctl restart wg-quick@wg0
```

**На боевом сервере до merge в main ничего не ставим.**

После merge:

```bash
cd /opt/vpn-manager
git pull
go build -o bin/vpnctl ./cmd/vpnctl
export VPN_DATA_DIR=/var/lib/vpn-manager
./bin/vpnctl user add bob
./bin/vpnctl server render | tee /etc/wireguard/wg0.conf
wg syncconf wg0 <(wg-quick strip wg0)
```

Установка: `sudo bash deploy/install.sh` (см. `deploy/`).

---

## Что читать в коде (порядок обучения)

1. `cmd/vpnctl/main.go` — вход
2. `cmd/vpnctl/cmd/init.go` — первая команда
3. `internal/wgkeys/keys.go` — крипто без внешнего `wg`
4. `internal/store/users.go` — SQL
5. `internal/configgen/configgen.go` — строковые шаблоны конфигов

---

## Деплой на боевой сервер (безопасно)

1. **Не** включай `ufw`, если он был выключен.
2. Порт **51821/udp** (не 51820, если занят).
3. Подсеть **10.99.0.0/24** — не пересекается с `docker0` (обычно 172.x).
4. Клиентам сначала только `--admin-only`.
5. Бэкап перед изменениями:

```bash
iptables-save > /root/iptables.backup.txt
docker ps > /root/docker-before.txt
```

---

## Дальнейшие шаги (когда освоишь базу)

- [ ] HTTP API (`net/http` или `chi`) — как FastAPI
- [ ] Telegram-бот для выдачи конфигов
- [ ] GitHub Actions: build + deploy по SSH
- [ ] Панель на TypeScript (отдельный фронт)

---

## Лицензия

Учебный проект. Используй на свой страх и риск; проверяй безопасность перед продакшеном.
