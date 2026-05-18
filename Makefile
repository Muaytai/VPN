.PHONY: build test run-init run-add tidy

build:
	go build -o bin/vpnctl ./cmd/vpnctl

test:
	go test ./...

tidy:
	go mod tidy

# Локальные примеры (Windows PowerShell: $env:VPN_DATA_DIR="./data")
run-init:
	go run ./cmd/vpnctl init --endpoint 127.0.0.1 --port 51821

run-add:
	go run ./cmd/vpnctl user add devuser
