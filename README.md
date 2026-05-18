# VPN Manager (учебный проект на Go)

Управление **WireGuard** через CLI: пользователи, ключи, генерация `.conf`.

> Мы **не** пишем свой VPN-протокол. WireGuard — в ядре Linux.  
> Наш код на Go — это «панель/скрипт», как админка на Python, только компилируется в один бинарник.

---

## Для кого этот README

Ты знаешь **Python + TypeScript**, Go видишь впервые. Ниже — карта соответствий и пошаговый старт.

---

## Python/TS → Go: шпаргалка

| Привычно (Py/TS) | В Go |
|------------------|------|
| `pip install` / `npm install` | `go mod download` |
| `requirements.txt` / `package.json` | `go.mod` + `go.sum` |
| пакет/модуль `myapp/users.py` | пакет `internal/store` (папка = пакет) |
| `class User` | `struct User` (без наследования) |
| `interface UserDTO` (TS) | `interface` — набор методов, реализуется неявно |
| `try/except`, `throw` | `if err != nil { return err }` — ошибки **значения**, не исключения |
| `async/await` | горутины + `chan` (позже; в MVP не нужны) |
| `if __name__ == "__main__"` | `cmd/vpnctl/main.go` → `func main()` |
| FastAPI / Click CLI | [Cobra](https://github.com/spf13/cobra) CLI |
| SQLAlchemy / Prisma | `database/sql` + SQLite (`internal/store`) |

**Важная идея Go:** явная обработка ошибок на каждом шаге. Это кажется многословным, зато путь ошибки всегда виден в коде.

```go
user, err := store.CreateUser(u)
if err != nil {
    return err  // как raise, но без stack trace по умолчанию
}
```

---

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

**По умолчанию: full-tunnel** — весь интернет через VPN (`AllowedIPs = 0.0.0.0/0`).  
Для доступа к Gemini и смены IP нужен `vpnctl server apply` на Linux-сервере (NAT + ip_forward).

Режим только до сервера: `vpnctl user add name --admin-only`

---

## Команды CLI

| Команда | Что делает |
|---------|------------|
| `vpnctl init` | Создаёт БД и ключи сервера |
| `vpnctl user add NAME` | Новый пользователь + `.conf` |
| `vpnctl user list` | Список пользователей |
| `vpnctl user revoke NAME` | Отзыв доступа |
| `vpnctl user regen NAME` | Пересоздать `.conf` (после смены режима) |
| `vpnctl server render` | `wg0.conf` с NAT (full-tunnel) |
| `vpnctl server apply` | ip_forward + записать wg0 + wg syncconf |
| `vpnctl server configure` | WAN / DNS / full-tunnel в БД |

---

## Репозиторий GitHub

Аккаунт: [github.com/Muaytai](https://github.com/Muaytai)

Репозиторий: [github.com/Muaytai/VPN](https://github.com/Muaytai/VPN)

```bash
git remote add origin https://github.com/Muaytai/VPN.git
```

Модуль Go в коде: `github.com/Muaytai/vpn-manager` (имя модуля в `go.mod`, путь репо может отличаться).

---

## Workflow: локально → GitHub → сервер

```
[Твой ПК]  git push   [GitHub]   git pull   [Сервер /opt/vpn-manager]
   │                                              │
   ├─ go test ./...                               ├─ go build
   ├─ vpnctl user add (dev)                       ├─ vpnctl init (один раз)
   └─ правки Go                                   ├─ vpnctl server apply
                                                  └─ user regen + переимпорт .conf
```

После merge на сервере:

```bash
cd /opt/vpn-manager
git pull
/usr/local/go/bin/go build -o bin/vpnctl ./cmd/vpnctl
export VPN_DATA_DIR=/opt/vpn-manager/data
iptables-save > /root/iptables.backup.txt
ip route | awk '/default/ {print $5}'   # запомни интерфейс: ens3, eth0, …
/opt/vpn-manager/bin/vpnctl server configure --wan ens3 --full-tunnel
/opt/vpn-manager/bin/vpnctl user regen alice
/opt/vpn-manager/bin/vpnctl server apply
```

На Windows: переимпортируй обновлённый `alice.conf` в WireGuard.

Установка: `sudo bash deploy/install.sh` (см. `deploy/`).

---

## Что читать в коде (порядок обучения)

1. `cmd/vpnctl/main.go` — вход
2. `cmd/vpnctl/cmd/init.go` — первая команда
3. `internal/wgkeys/keys.go` — крипто без внешнего `wg`
4. `internal/store/users.go` — SQL
5. `internal/configgen/configgen.go` — строковые шаблоны конфигов

---

## Full-tunnel и Gemini

1. Сервер в NL/EU — IP датацентра; Gemini **может** заработать, но Google иногда блокирует VPS.
2. Проверка IP: https://ifconfig.me — должен быть IP сервера `188.137.180.149`.
3. `server apply` добавляет NAT — **проверь Docker-проект** после apply.
4. Порт **51821/udp** открыт у хостера.
5. Бэкап перед `server apply`:

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
